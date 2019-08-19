package docker

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/registry"
)

// Run will start a container, copy all In-Toto metadata in /in-toto
// then run in-toto-verification
func Run(image, layout, key, linksDir string, targetFiles []string) error {
	ctx := context.Background()
	cli, err := initializeDockerCli()
	if err != nil {
		return err
	}

	cfg := &container.Config{
		Image:      image,
		WorkingDir: "/in-toto",
		Cmd:        []string{"in-toto-verify", "--layout", "layout.template", "--layout-keys", "key.pub", "--link-dir", "/in-toto", "--verbose"},
		//Cmd:          []string{"sleep", "500"},
		AttachStderr: true,
		AttachStdout: true,
	}
	resp, err := cli.Client().ContainerCreate(ctx, cfg, &container.HostConfig{}, nil, "ver")
	switch {
	case client.IsErrNotFound(err):
		fmt.Fprintf(cli.Err(), "Unable to find image '%s' locally\n", image)
		if err := pullImage(ctx, cli, image); err != nil {
			return err
		}
		if resp, err = cli.Client().ContainerCreate(ctx, cfg, &container.HostConfig{}, nil, ""); err != nil {
			return fmt.Errorf("cannot create container: %v", err)
		}
	case err != nil:
		return fmt.Errorf("cannot create container: %v", err)
	}

	defer cli.Client().ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

	files, err := buildFileMap(layout, key, linksDir, targetFiles)
	if err != nil {
		return err
	}

	arch, err := generateArchive(files)
	if err != nil {
		return err
	}

	copyOpts := types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
	}

	err = cli.Client().CopyToContainer(ctx, resp.ID, "/", arch, copyOpts)
	if err != nil {
		return err
	}

	attach, err := cli.Client().ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return fmt.Errorf("unable to retrieve logs: %v", err)
	}

	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr

	go func() {
		defer attach.Close()
		for {
			_, err := stdcopy.StdCopy(stdout, stderr, attach.Reader)
			if err != nil {
				break
			}
		}
	}()

	statusc, errc := cli.Client().ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	if err = cli.Client().ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("cannot start container: %v", err)
	}
	select {
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("error in container: %v", err)
		}
	case s := <-statusc:
		if s.Error != nil {
			return fmt.Errorf("container exit code %v: %v", s.StatusCode, err)
		}
	}
	return nil
}

func pullImage(ctx context.Context, cli command.Cli, image string) error {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return err
	}
	repo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}

	authCfg := command.ResolveAuthConfig(ctx, cli, repo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authCfg)
	if err != nil {
		return err
	}

	opts := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}

	responseBody, err := cli.Client().ImagePull(ctx, image, opts)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesStream(responseBody, cli.Out(), cli.Out().FD(), false, nil)
}

func initializeDockerCli() (command.Cli, error) {
	cli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}

	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return nil, err
	}
	return cli, nil
}

// TODO - allow passing multiple signing keys
func generateArchive(files map[string][]byte) (io.Reader, error) {
	r, w := io.Pipe()
	tw := tar.NewWriter(w)

	go func() {
		for p, c := range files {
			fmt.Printf("copying file %v in container for verification...\n", p)
			hdr := &tar.Header{
				Name: p,
				Mode: 0644,
				Size: int64(len(c)),
			}

			tw.WriteHeader(hdr)
			tw.Write(c)
		}
		w.Close()
	}()

	return r, nil
}

func buildFileMap(layout, key, linksDir string, targeFiles []string) (map[string][]byte, error) {
	files := make(map[string][]byte)

	lc, err := ioutil.ReadFile(layout)
	if err != nil {
		return nil, err
	}
	files[layoutPath] = lc

	kc, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}
	files[keyPath] = kc

	f, err := ioutil.ReadDir(linksDir)
	if err != nil {
		return nil, err
	}
	for _, li := range f {
		if !strings.Contains(li.Name(), ".link") {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(linksDir, li.Name()))
		if err != nil {
			return nil, err
		}

		files[filepath.Join("in-toto", li.Name())] = b
	}

	for _, fi := range targeFiles {
		by, err := ioutil.ReadFile(fi)
		if err != nil {
			return nil, err
		}

		files[filepath.Join("in-toto", path.Base(fi))] = by
	}

	return files, nil
}

const (
	layoutPath = "/in-toto/layout.template"
	keyPath    = "/in-toto/key.pub"
)
