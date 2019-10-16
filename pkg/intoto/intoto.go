package intoto

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	canonicaljson "github.com/docker/go/canonical/json"
	"github.com/in-toto/in-toto-golang/in_toto"
	log "github.com/sirupsen/logrus"
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

	log.Infof("Verification succeeded for layout %v", layout)
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

// validateRSAPubKey checks if a passed key is a valid RSA public key.
func validateRSAPubKey(key in_toto.Key) error {
	if key.KeyType != "rsa" {
		return fmt.Errorf("invalid KeyType for key '%s': should be 'rsa', got"+
			" '%s'", key.KeyId, key.KeyType)
	}
	if key.Scheme != "rsassa-pss-sha256" {
		return fmt.Errorf("invalid scheme for key '%s': should be "+
			"'rsassa-pss-sha256', got: '%s'", key.KeyId, key.Scheme)
	}
	if err := validatePubKey(key); err != nil {
		return err
	}
	return nil
}

// validatePubKey is a general function to validate if a key is a valid public key.
func validatePubKey(key in_toto.Key) error {
	if err := validateHexString(key.KeyId); err != nil {
		return fmt.Errorf("keyid: %s", err.Error())
	}
	if key.KeyVal.Private != "" {
		return fmt.Errorf("in key '%s': private key found", key.KeyId)
	}
	if key.KeyVal.Public == "" {
		return fmt.Errorf("in key '%s': public key cannot be empty", key.KeyId)
	}
	return nil
}

// validateHexString is used to validate that a string passed to it contains
// only valid hexadecimal characters.
func validateHexString(str string) error {
	formatCheck, _ := regexp.MatchString("^[a-fA-F0-9]+$", str)
	if !formatCheck {
		return fmt.Errorf("'%s' is not a valid hex string", str)
	}
	return nil
}

// validateStep ensures that a passed step is valid and matches the
// necessary format of an step.
func validateStep(step in_toto.Step) error {
	if err := validateSupplyChainItem(step.SupplyChainItem); err != nil {
		return fmt.Errorf("step %s", err.Error())
	}
	if step.Type != "step" {
		return fmt.Errorf("invalid Type value for step '%s': should be 'step'",
			step.SupplyChainItem.Name)
	}
	for _, keyID := range step.PubKeys {
		if err := validateHexString(keyID); err != nil {
			return fmt.Errorf("in step '%s', keyid: %s",
				step.SupplyChainItem.Name, err.Error())
		}
	}
	return nil
}

// validateSupplyChainItem is used to validate the common elements found in both
// steps and inspections. Here, the function primarily ensures that the name of
// a supply chain item isn't empty.
func validateSupplyChainItem(item in_toto.SupplyChainItem) error {
	if item.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if err := validateSliceOfArtifactRules(item.ExpectedMaterials); err != nil {
		return fmt.Errorf("invalid material rule: %s", err)
	}
	if err := validateSliceOfArtifactRules(item.ExpectedProducts); err != nil {
		return fmt.Errorf("invalid product rule: %s", err)
	}
	return nil
}

// validateArtifactRule calls UnpackRule to validate that the passed rule conforms
// with any of the available rule formats.
func validateArtifactRule(rule []string) error {
	if _, err := in_toto.UnpackRule(rule); err != nil {
		return err
	}
	return nil
}

// validateSliceOfArtifactRules iterates over passed rules to validate them.
func validateSliceOfArtifactRules(rules [][]string) error {
	for _, rule := range rules {
		if err := validateArtifactRule(rule); err != nil {
			return err
		}
	}
	return nil
}

// GetLayout returns an In-Toto layout given a file path
func getLayout(layout string) (*in_toto.Layout, error) {
	var mb in_toto.Metablock
	if err := mb.Load(layout); err != nil {
		return nil, fmt.Errorf("cannot load layout from file file %v: %v", layout, err)
	}
	l := mb.Signed.(in_toto.Layout)
	return &l, nil
}

// ValidateFromPath validates a layout given a path
func ValidateFromPath(p string) error {
	l, err := getLayout(p)
	if err != nil {
		return err
	}
	return ValidateLayout(*l)
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
