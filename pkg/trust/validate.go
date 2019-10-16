package trust

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/engineerd/signy/pkg/docker"
	"github.com/engineerd/signy/pkg/intoto"
	"github.com/engineerd/signy/pkg/tuf"
)

// Validate performs all trust validations
func Validate(ref, trustServer, tlscacert, trustDir, verificationImage string, targets []string, keep bool) error {
	err := tuf.VerifyCNABTrust(ref, trustServer, tlscacert, trustDir)
	if err != nil {
		return err
	}

	target, _, err := tuf.GetTargetAndSHA(ref, trustServer, tlscacert, trustDir)
	if err != nil {
		return err
	}

	m := &intoto.Metadata{}
	err = json.Unmarshal(*target.Custom, m)
	if err != nil {
		return err
	}

	verificationDir, err := ioutil.TempDir(os.TempDir(), "intoto-verification")
	if err != nil {
		return err
	}
	if !keep {
		defer func() {
			os.RemoveAll(verificationDir)
			os.Remove(verificationDir)
		}()
	}

	fmt.Printf("\nWriting In-Toto metadata files into %v", verificationDir)
	err = intoto.WriteMetadataFiles(m, verificationDir)
	if err != nil {
		return err
	}

	return docker.Run(verificationImage, filepath.Join(verificationDir, intoto.LayoutDefaultName), filepath.Join(verificationDir, intoto.KeyDefaultName), verificationDir, targets)
}
