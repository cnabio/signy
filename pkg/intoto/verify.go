package intoto

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	canonicaljson "github.com/docker/go/canonical/json"
	"github.com/in-toto/in-toto-golang/in_toto"
)

const (
	LayoutDefaultName = "layout.template"
	KeyDefaultName    = "key.pub"
)

// Verify performs the in-toto validation steps
func Verify(layout, linkDir string, layoutKeyPaths ...string) error {
	layoutKeys := make(map[string]in_toto.Key)
	for _, p := range layoutKeyPaths {
		var k in_toto.Key
		if err := k.LoadPublicKey(p); err != nil {
			return fmt.Errorf("cannot load layout public key %v: %v", p, err)
		}
		layoutKeys[k.KeyId] = k
	}

	var mb in_toto.Metablock
	if err := mb.Load(layout); err != nil {
		return fmt.Errorf("cannot load layout from file file %v: %v", layout, err)
	}

	if err := ValidateLayout(mb.Signed.(in_toto.Layout)); err != nil {
		return fmt.Errorf("invalid metatada found: %v", err)
	}

	if _, err := in_toto.InTotoVerifyWithDirectory(mb, layoutKeys, linkDir, linkDir, "", make(map[string]string)); err != nil {
		return fmt.Errorf("failed verification: %v", err)
	}

	fmt.Printf("Verification succeeded for layout %v.\n", layout)
	return nil
}

// ValidateLayout is a function used to ensure that a passed item of type Layout
// matches the necessary format.
func ValidateLayout(layout in_toto.Layout) error {
	if layout.Type != "layout" {
		return fmt.Errorf("invalid Type value for layout: should be 'layout'")
	}

	if _, err := time.Parse(in_toto.ISO8601DateSchema, layout.Expires); err != nil {
		return fmt.Errorf("expiry time parsed incorrectly - date either" +
			" invalid or of incorrect format")
	}

	for keyID, key := range layout.Keys {
		if key.KeyId != keyID {
			return fmt.Errorf("invalid key found")
		}
		if err := validateRSAPubKey(key); err != nil {
			return err
		}
	}

	var namesSeen = make(map[string]bool)
	for _, step := range layout.Steps {
		if namesSeen[step.Name] {
			return fmt.Errorf("non unique step or inspection name found")
		}
		namesSeen[step.Name] = true
		if err := validateStep(step); err != nil {
			return err
		}
	}

	for _, inspection := range layout.Inspect {
		if namesSeen[inspection.Name] {
			return fmt.Errorf("non unique step or inspection name found")
		}
		namesSeen[inspection.Name] = true
	}
	return nil
}

// GetMetadataRawMessage takes In-Toto metadata and returns a canonical RawMessage
// that can be stored in the TUF targets custom field.
//
// TODO: layout signing key should not be passed by the library.
// Layouts should be signed with the targets key used to sign the TUF collection.
func GetMetadataRawMessage(layout string, linkDir string, layoutKey string) (canonicaljson.RawMessage, error) {
	k, err := ioutil.ReadFile(layoutKey)
	if err != nil {
		return nil, fmt.Errorf("cannot get canonical JSON from file %v: %v", layoutKey, err)
	}

	l, err := ioutil.ReadFile(layout)
	if err != nil {
		return nil, fmt.Errorf("cannot get canonical JSON from file %v: %v", layout, err)
	}

	links := make(map[string][]byte)
	files, err := ioutil.ReadDir(linkDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read links directory %v: %v", linkDir, err)
	}
	for _, f := range files {
		// TODO - Radu M
		//
		// robust check if file is actually a link
		if !strings.Contains(f.Name(), ".link") {
			continue
		}
		b, err := ioutil.ReadFile(filepath.Join(linkDir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("cannot get canonical JSON from file %v: %v", f.Name(), err)
		}
		links[f.Name()] = b
	}

	m := &Metadata{
		Key:    k,
		Layout: l,
		Links:  links,
	}

	raw, err := canonicaljson.MarshalCanonical(m)
	if err != nil {
		return nil, fmt.Errorf("cannot encode in-toto metadata into canonical json %v: %v", m, err)
	}

	return canonicaljson.RawMessage(raw), nil
}

// Metadata represents the In-Toto metadata stored in TUF.
// All fields are represented as []byte in order to be stored in the Custom field for TUF metadata.
type Metadata struct {
	// TODO: remove this once the TUF targets key is used to sign the root layout
	Key    []byte            `json:"key"`
	Layout []byte            `json:"layout"`
	Links  map[string][]byte `json:"links"`
}

// WriteMetadataFiles writes the content of a metadata object into files in a directory
func WriteMetadataFiles(m *Metadata, dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(abs, LayoutDefaultName), m.Layout, 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(abs, KeyDefaultName), m.Key, 0644)
	if err != nil {
		return err
	}

	for n, c := range m.Links {
		err = ioutil.WriteFile(filepath.Join(abs, n), c, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
