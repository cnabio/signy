package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"

	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
	dockerClient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type pushCmd struct {
	ref       string
	thick     bool
	localFile string

	intoto     bool
	verifyOnOS bool
	pushImage  string

	intotoLayout    string
	intotoLayoutKey string
	intotoLinkDir   string

	registryCredentials string
	registryUser        string

	notaryServer string
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

	cmd.Flags().StringVarP(&push.pushImage, "image", "i", "", "container image to push")

	cmd.Flags().StringVarP(&push.intotoLayout, "intotoLayout", "", "intoto/root.layout", "intotoLayout")
	cmd.Flags().StringVarP(&push.intotoLayoutKey, "intotoLayoutKey", "", "intoto/alice.pub", "intotoLayoutKey")
	cmd.Flags().StringVarP(&push.intotoLinkDir, "intotoLinkDir", "", "intoto/", "intotoLinkDir")

	cmd.Flags().StringVarP(&push.registryUser, "registryUser", "", viper.GetString("PUSH_REGISTRY_USER"), "docker registry user")
	cmd.Flags().StringVarP(&push.registryCredentials, "registryCredentials", "", viper.GetString("PUSH_REGISTRY_CREDENTIALS"), "docker registry credentials (api key or password)")

	cmd.Flags().StringVarP(&push.notaryServer, "notaryServer", "", viper.GetString("PUSH_NOTARY_SERVER"), "notary server")

	return cmd
}

type Metadata struct {
	// TODO: remove this once the TUF targets key is used to sign the root layout
	Key    []byte            `json:"key"`
	Layout []byte            `json:"layout"`
	Links  map[string][]byte `json:"links"`
}

//cd /Users/scottbuckel/Desktop/push-with-intoto/push-with-intoto && make bootstrap build && cd bin && ./push-with-intoto push -i sebbyii/testimage:test && cd ..
func (v *pushCmd) run() error {

	if v.pushImage == "" {
		return fmt.Errorf("Must specify an image for push")
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
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	ctx := context.Background()

	//push the image
	resp, err := cli.ImagePush(ctx, v.pushImage, types.ImagePushOptions{RegistryAuth: authStr})
	defer resp.Close()

	//for debugging, or else you cant see if wrong pw or whatnot
	//TODO: How to see this info all the time?
	//io.Copy(os.Stdout, resp)

	//get the result of push, this is weird because it requires getting the aux. value of the response
	pushResult, err := parseDockerDaemonJsonMessages(resp)
	if err != nil {
		return err
	}

	//again, just for debugging
	//io.Copy(os.Stdout, resp)

	if v.intotoLayoutKey == "" || v.intotoLayout == "" || v.intotoLinkDir == "" {
		return fmt.Errorf("required in-toto metadata not found")
	}
	log.Infof("Adding In-Toto layout and links metadata to TUF")
	err = intoto.ValidateFromPath(v.intotoLayout)
	if err != nil {
		return fmt.Errorf("validation for in-toto metadata failed: %v", err)
	}
	custom, err := intoto.GetMetadataRawMessage(v.intotoLayout, v.intotoLinkDir, v.intotoLayoutKey)
	if err != nil {
		return fmt.Errorf("cannot get metadata message: %v", err)
	}

	//Sign and publish and get a target back
	target, err := tuf.SignAndPublishWithTarget(trustDir, v.notaryServer, v.pushImage, pushResult, "root-ca.crt", "", "20s", &custom)
	if err != nil {
		fmt.Errorf("cannot sign and publish trust data: %v", err)

	}

	log.Infof("Pushed trust data for %v: %v\n", v.pushImage, "with sha256 of "+string(target.Hashes["sha256"]))

	return nil
}

func parseDockerDaemonJsonMessages(r io.Reader) (types.PushResult, error) {
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
