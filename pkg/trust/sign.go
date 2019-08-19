package trust

import (
	"encoding/hex"
	"fmt"

	"github.com/engineerd/signy/pkg/cnab"
	"github.com/engineerd/signy/pkg/intoto"
	"github.com/engineerd/signy/pkg/tuf"
)

// SignAndPublish takes a CNAB bundle, pushes the signature and metadata to a trust server, then pushes the bundle
func SignAndPublish(ref, layout, linkDir, layoutKey, trustDir, trustServer, file, tlscacert string) error {
	err := intoto.ValidateFromPath(layout)
	if err != nil {
		return fmt.Errorf("validation for in-toto metadata failed: %v", err)
	}
	r, err := intoto.GetMetadataRawMessage(layout, linkDir, layoutKey)
	if err != nil {
		return fmt.Errorf("cannot get metadata message: %v", err)
	}

	fmt.Printf("\nAdding In-Toto layout and links metadata to TUF")

	target, err := tuf.SignAndPublish(trustDir, trustServer, ref, file, tlscacert, "", &r)
	if err != nil {
		return fmt.Errorf("cannot sign and publish trust data: %v", err)
	}
	fmt.Printf("\nPushed trust data for %v: %v to server %v\n", ref, hex.EncodeToString(target.Hashes["sha256"]), trustServer)
	return cnab.Push(file, ref)
}
