package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//cd /Users/scottbuckel/Desktop/push-with-intoto/push-with-intoto && make bootstrap build && cd bin && ./push-with-intoto pull -i sebbyii/testimage:test && cd ..
type pullCmd struct {
	ref       string
	thick     bool
	localFile string

	intoto     bool
	verifyOnOS bool
	pullImage  string

	intotoLayout    string
	intotoLayoutKey string
	intotoLinkDir   string

	notaryServer string
}

func newPullCmd() *cobra.Command {
	const pullDesc = `
Pull to docker and notary with trust data`
	pull := pullCmd{}
	cmd := &cobra.Command{
		Use:   "pull [target reference]",
		Short: "Pull",
		Long:  pullDesc,
		//Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			//		push.ref = args[0]
			return pull.run()
		},
	}

	cmd.Flags().StringVarP(&pull.pullImage, "image", "i", "", "container image to pull")

	cmd.Flags().StringVarP(&pull.notaryServer, "notaryServer", "", viper.GetString("PUSH_NOTARY_SERVER"), "notary server")

	return cmd
}

//cd /Users/scottbuckel/Desktop/push-with-intoto/push-with-intoto && make bootstrap build && cd bin && ./push-with-intoto pull -i sebbyii/testimage:test && cd ..
func (v *pullCmd) run() error {

	if v.pullImage == "" {
		return fmt.Errorf("Must specify an image for pull")
	}

	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	//pull the actual image
	reader, err := cli.ImagePull(ctx, v.pullImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	//there has to be a better way do do this, basically we inspect the image we just puleld, that image has a few digests (for example, if an image was tagged multiple times)
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

	//pull the data from notary
	target, trustedSHA, err := tuf.GetTargetAndSHAString(v.pullImage, v.notaryServer, "root-ca.crt", trustDir, "20s")
	if err != nil {
		return err
	}

	if pulledSHA == trustedSHA {
		log.Infof("Pulled SHA matches trusted notary SHA: SHA256: %v matches %v", pulledSHA, trustedSHA)
	} else {
		return fmt.Errorf("Our pulled image doesn't match notary!!! ")
	}

	verificationDir, err := intoto.GetVerificationDir(target)

	if err != nil {
		return err
	}

	spew.Dump(verificationDir)

	//temporary to get compiler not to bitch
	if target == nil {
		spew.Dump(target)
	}

	spew.Dump(trustedSHA)

	//temp so i dont get errors
	if reader == nil {
		spew.Dump(reader)
	}

	return intoto.VerifyInContainer(target, []byte(v.pullImage), "sebbyii/signy-intoto-verifier", logLevel)

}
