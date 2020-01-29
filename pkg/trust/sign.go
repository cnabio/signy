package trust

import (
	"encoding/hex"
	"fmt"

	"github.com/cnabio/signy/pkg/cnab"
	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
	log "github.com/sirupsen/logrus"
)

// SignAndPublish takes a CNAB bundle, pushes the signature and metadata to a trust server, then pushes the bundle
func SignAndPublish(ref, layout, linkDir, layoutKey, trustDir, trustServer, file, tlscacert, timeout string) error {
	err := intoto.ValidateFromPath(layout)
	if err != nil {
		return fmt.Errorf("validation for in-toto metadata failed: %v", err)
	}
	r, err := intoto.GetMetadataRawMessage(layout, linkDir, layoutKey)
	if err != nil {
		return fmt.Errorf("cannot get metadata message: %v", err)
	}

	log.Infof("Adding In-Toto layout and links metadata to TUF")

	target, err := tuf.SignAndPublish(trustDir, trustServer, ref, file, tlscacert, "", timeout, &r)
	if err != nil {
		return fmt.Errorf("cannot sign and publish trust data: %v", err)
	}
	log.Infof("Pushed trust data for %v: %v to server %v", ref, hex.EncodeToString(target.Hashes["sha256"]), trustServer)
	return cnab.Push(file, ref)
}
