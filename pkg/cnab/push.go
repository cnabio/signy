package cnab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

// Push pushes a bundle to an OCI registry
func Push(bundleFile, ref string) error {
	buf, err := ioutil.ReadFile(bundleFile)
	if err != nil {
		return fmt.Errorf("cannot read bundle file: %v", err)
	}

	var b bundle.Bundle
	if err = json.Unmarshal(buf, &b); err != nil {
		return err
	}

	resolver := createResolver(nil)
	n, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	fixupOpts := []remotes.FixupOption{
		remotes.WithEventCallback(displayEvent),
		remotes.WithInvocationImagePlatforms(nil),
		// we explicitly DO NOT want to update the bundle file after the trust data has been pushed
		// remotes.WithAutoBundleUpdate(),
		remotes.WithPushImages(cli, os.Stdout),
		remotes.WithComponentImagePlatforms(nil),
	}

	relocationMap, err := remotes.FixupBundle(context.Background(), &b, n, resolver, fixupOpts...)
	if err != nil {
		return err
	}

	log.Infof("Generated relocation map: %#v", relocationMap)
	d, err := remotes.Push(context.Background(), &b, relocationMap, n, resolver, true)
	if err != nil {
		return err
	}

	log.Infof("Pushed successfully, with digest %q\n", d.Digest)
	return nil
}
