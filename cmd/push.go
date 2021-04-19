package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/theupdateframework/notary"

	dockerClient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
)

type pushCmd struct {
	pushImage string

	layout string
	// TODO: figure out a way to pass layout root key to TUF (not in the custom object)
	layoutKey string
	linkDir   string

	registryCredentials string
	registryUser        string
}

func newPushCmd() *cobra.Command {
	const pushDesc = `
Push to docker and notary with trust data`
	push := pushCmd{}
	cmd := &cobra.Command{
		Use:   "push [target reference]",
		Short: "Push",
		Long:  pushDesc,
		//Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			//		push.ref = args[0]
			return push.run()
		},
	}

	//need to set this to automatically check for env variables for viper.get
	viper.AutomaticEnv()

	cmd.Flags().StringVarP(&push.pushImage, "image", "i", "", "container image to push (must be built on your local system)")
	cmd.Flags().StringVarP(&push.layout, "layout", "", "intoto/root.layout", "Path to the in-toto root layout file")
	cmd.Flags().StringVarP(&push.linkDir, "links", "", "intoto/", "Path to the in-toto links directory")
	cmd.Flags().StringVarP(&push.layoutKey, "layout-key", "", "intoto/root.pub", "Path to the in-toto root layout public keys")
	cmd.Flags().StringVarP(&push.registryUser, "registryUser", "", viper.GetString("PUSH_REGISTRY_USER"), "docker registry user")
	cmd.Flags().StringVarP(&push.registryCredentials, "registryCredentials", "", viper.GetString("PUSH_REGISTRY_CREDENTIALS"), "docker registry credentials (api key or password)")

	return cmd
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

	//for debugging, or else you cant see if wrong pw or whatnot
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
