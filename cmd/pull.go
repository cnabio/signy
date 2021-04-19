package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cnabio/signy/pkg/docker"
	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type pullCmd struct {
	pullImage         string
	verificationImage string
}

func newPullCmd() *cobra.Command {
	const pullDesc = `
		Pull to docker and notary with trust data`
	pull := pullCmd{}
	cmd := &cobra.Command{
		Use:   "pull [target reference]",
		Short: "Pull",
		Long:  pullDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pull.run()
		},
	}

	cmd.Flags().StringVarP(&pull.pullImage, "image", "i", "", "container image to pull")
	cmd.Flags().StringVarP(&pull.verificationImage, "intotoVerificationImage", "v", docker.VerificationImage, "container image to run the in-toto verification")

	return cmd
}

func (v *pullCmd) run() error {

	if v.pullImage == "" {
		return fmt.Errorf("Must specify an image for pull")
	}

	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Couldn't initialize dockerClient")
	}

	//pull the image from the repository
	log.Infof("Pulling image %v from registry", v.pullImage)

	_, err = cli.ImagePull(ctx, v.pullImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Couldnt pull image %v", err)
	}

	//there has to be a better way do do this, we inspect the image we just pulled, that image has a few digests (for example, if an image was tagged multiple times)
	imageDigests, _, err := cli.ImageInspectWithRaw(ctx, v.pullImage)
	pulledSHA := ""
	for _, element := range imageDigests.RepoDigests {

		//remove the tag, since we have only digest now (image@sha256:)
		parts := strings.Split(v.pullImage, ":")

		if strings.Contains(element, parts[0]) {
			//remove the image:@sha256, return only the actual sha
			pulledSHA = strings.Split(element, ":")[1]
		}
	}

	log.Infof("Successfully pulled image %v", v.pullImage)

	//pull the data from notary
	target, trustedSHA, err := tuf.GetTargetAndSHA(v.pullImage, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return err
	}

	if pulledSHA == trustedSHA {
		log.Infof("Pulled SHA matches TUF SHA: SHA256: %v matches %v", pulledSHA, trustedSHA)
	} else {
		return fmt.Errorf("Pulled image digest doesn't match TUF SHA! Pulled SHA: %v doesn't match TUF SHA: %v ", pulledSHA, trustedSHA)
	}

	if target.Custom == nil {
		return fmt.Errorf("Error: TUF server doesn't have the custom field filled with in-toto metadata.")
	}

	/*
		TODO: Allow other verifications like `Signy verify` does, also fail better when RuleVerificationError happen
	*/
	return intoto.VerifyInContainer(target, []byte(v.pullImage), v.verificationImage, logLevel)
}
