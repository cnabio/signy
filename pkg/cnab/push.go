package cnab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
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

	relocationMap, err := remotes.FixupBundle(context.Background(), &b, n, resolver, remotes.WithEventCallback(displayEvent),
		remotes.WithInvocationImagePlatforms(nil),
		remotes.WithAutoBundleUpdate(),
		remotes.WithComponentImagePlatforms(nil))
	if err != nil {
		return err
	}

	fmt.Printf("\nGenerated relocation map: %#v", relocationMap)
	d, err := remotes.Push(context.Background(), &b, relocationMap, n, resolver, true)
	if err != nil {
		return err
	}

	fmt.Printf("\nPushed successfully, with digest %q\n", d.Digest)
	return nil
}
