package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theupdateframework/notary"

	"github.com/cnabio/signy/pkg/docker"
	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
)

func buildImageCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Image commands",
		Long:  "Commands for working with images.",
	}

	cmd.AddCommand(buildImagePullCommand())
	cmd.AddCommand(buildImagePushCommand())
	return cmd
}

func buildImagePushCommand() *cobra.Command {
	push := pushCmd{}
	cmd := &cobra.Command{
		Use:   "push [target reference]",
		Short: "Pushes an image to a registry and trust data to TUF",
		Long:  "Pushes an image to a registry and gets it's digest. After it's pushed, it pushes the digest to TUF alongside it's in-toto metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return push.run()
		},
	}

	//need to set this to automatically check for env variables for viper.get
	viper.AutomaticEnv()

	cmd.Flags().StringVarP(&push.pushImage, "image", "i", "", "container image to push (must be built on your local system)")
	cmd.Flags().StringVarP(&push.layout, "layout", "", "intoto/root.layout", "Path to the in-toto root layout file")
	cmd.Flags().StringVarP(&push.linkDir, "links", "", "intoto/", "Path to the in-toto links directory")
	cmd.Flags().StringVarP(&push.layoutKey, "layout-key", "", "intoto/root.pub", "Path to the in-toto root layout public keys")
	cmd.Flags().StringVarP(&push.registryUser, "registryUser", "", viper.GetString("PUSH_REGISTRY_USER"), "docker registry user, also uses the PUSH_REGISTRY_USER environment variable")
	cmd.Flags().StringVarP(&push.registryCredentials, "registryCredentials", "", viper.GetString("PUSH_REGISTRY_CREDENTIALS"), "docker registry credentials (api key or password), uses the PUSH_REGISTRY_CREDENTIALS environment variable")

	return cmd
}

func buildImagePullCommand() *cobra.Command {

	pull := pullCmd{}
	cmd := &cobra.Command{
		Use:   "pull [target reference]",
		Short: "Pulls an image from a registry and trust data from TUF and verifies it",
		Long:  "Pulls an image from a registry. After it's pulled, it compares it's digest with what was stored in TUF and then verifies its in-toto metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pull.run()
		},
	}

	cmd.Flags().StringVarP(&pull.pullImage, "image", "i", "", "container image to pull")
	cmd.Flags().StringVarP(&pull.verificationImage, "intotoVerificationImage", "v", docker.VerificationImage, "container image to run the in-toto verification")

	return cmd

}

type pullCmd struct {
	pullImage         string
	verificationImage string
}

type pushCmd struct {
	pushImage string

	layout string
	// TODO: figure out a way to pass layout root key to TUF (not in the custom object)
	layoutKey string
	linkDir   string

	registryCredentials string
	registryUser        string
}

func (v *pullCmd) run() error {

	if v.pullImage == "" {
		return fmt.Errorf("Must specify an image for pull")
	}

	//if the user is using the default verification image, check that signy was built with a tag for that image
	if strings.HasSuffix(v.verificationImage, ":") {
		return fmt.Errorf("Tag not specfied for the verification image. If using the default image, maybe you didn't compile with TAG= set")
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
	if err != nil {
		return err
	}

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
		return fmt.Errorf("Error: TUF server doesn't have the custom field filled with in-toto metadata")
	}

	/*
		TODO: Allow other verifications like `Signy verify` does, also fail better when RuleVerificationError happen
			//return intoto.VerifyInContainer(target, []byte(v.pullImage), v.verificationImage, logLevel)
	*/
	return intoto.VerifyOnOS(target, []byte(v.pullImage))
}

func (v *pushCmd) run() error {

	if v.pushImage == "" {
		return fmt.Errorf("Must specify an image for push")
	}
	if v.layout == "" || v.linkDir == "" || v.layoutKey == "" {
		return fmt.Errorf("Required in-toto metadata not found")
	}

	if intoto.ValidateFromPath(v.layout) != nil {
		return fmt.Errorf("validation for in-toto metadata failed")
	}

	//set up our docker client
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	//setup auth to docker repo
	authConfig := types.AuthConfig{
		Username: v.registryUser,
		Password: v.registryCredentials,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	ctx := context.Background()

	log.Infof("Pushing image %v to registry", v.pushImage)

	//push the image
	resp, err := cli.ImagePush(ctx, v.pushImage, types.ImagePushOptions{RegistryAuth: authStr})
	defer resp.Close()
	if err != nil {
		return fmt.Errorf("cannot push image to repository: %v", err)
	}

	//for debugging, or else you cant see if wrong pw
	//TODO: How to see this info all the time? if you consume it, it's no longer usable in the future
	//io.Copy(os.Stdout, resp)

	//get the result of push, this is weird because it requires getting the aux. value of the response
	pushResult, err := parseDockerDaemonJSONMessages(resp)
	if err != nil {
		return err
	}

	log.Infof("Image successfully pushed: {tag, sha, size} %v", pushResult)

	log.Infof("Adding In-Toto layout and links metadata to TUF")

	//get the json message we'll be adding to the custom field
	custom, err := intoto.GetMetadataRawMessage(v.layout, v.linkDir, v.layoutKey)
	if err != nil {
		return fmt.Errorf("cannot get metadata message: %v", err)
	}

	//Sign and publish and get a target back
	target, err := tuf.SignAndPublishWithImagePushResult(trustDir, trustServer, v.pushImage, pushResult, tlscacert, "", timeout, &custom)
	if err != nil {
		return fmt.Errorf("cannot sign and publish trust data: %v", err)
	}

	log.Infof("Pushed trust data for %v: %v ", v.pushImage, hex.EncodeToString(target.Hashes[notary.SHA256]))

	return nil
}

//the docker daemon responds with a lot of messages. we're only interested in the response with the aux field, which contains the digest
func parseDockerDaemonJSONMessages(r io.Reader) (types.PushResult, error) {
	var result types.PushResult

	decoder := json.NewDecoder(r)
	for {
		var jsonMessage jsonmessage.JSONMessage

		if err := decoder.Decode(&jsonMessage); err != nil {
			if err == io.EOF {
				break
			}
			return result, err
		}
		if err := jsonMessage.Error; err != nil {
			return result, err
		}
		if jsonMessage.Aux != nil {
			var r types.PushResult
			if err := json.Unmarshal(*jsonMessage.Aux, &r); err != nil {
				logrus.Warnf("Failed to unmarshal aux message. Cause: %s", err)
			} else {
				result = r
			}
		}
	}
	return result, nil
}
