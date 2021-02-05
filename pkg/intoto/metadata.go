package intoto

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	canonicaljson "github.com/docker/go/canonical/json"
)

// Metadata points to root layout, its public keys, and/or links
type Metadata struct {
	Data       []byte   `json:"data"`
	PublicKeys []string `json:"pubkeys"` // filenames
	Links      []string `json:"links"`   // filenames
}

// Custom is a generic structure that contains in-toto Metadata
type Custom struct {
	InToto Metadata `json:"intoto"`
}

// PublicKeys is a map from the GUN-qualified filename
// (e.g., "example.com/example-org/example-bundle/in-toto-pubkeys/keyid.pub")
// to a Custom struct
type PublicKeys map[string]Custom

// RootLayout maps a GUN-qualified filename
// (e.g., "example.com/example-org/example-bundle/in-toto-metadata/root.layout")
// to a Custom struct
type RootLayout struct {
	Filename string
	Custom   Custom
}

// Links is a map from the GUN-qualified filename
// (e.g., "example.com/example-org/example-bundle/in-toto-metadata/DIGEST/step.link")
// to a Custom struct
type Links map[string]Custom

// GetRawMessage transforms a Custom struct into a raw Canonical JSON message
func (custom *Custom) GetRawMessage() (canonicaljson.RawMessage, error) {
	marshalled, err := canonicaljson.MarshalCanonical(custom)
	if err != nil {
		return nil, fmt.Errorf("cannot encode custom metadata into canonical json: %v", err)
	}
	return canonicaljson.RawMessage(marshalled), nil
}

// GetPublicKeys reads public keys off disk into a PublicKeys map
func GetPublicKeys(gun string, filenames ...string) (PublicKeys, error) {
	publicKeys := make(PublicKeys)

	for _, filename := range filenames {
		if !strings.HasSuffix(filename, ".pub") {
			return nil, fmt.Errorf("%s does not have a .pub suffix", filename)
		}

		data, err := readFile(filename)
		if err != nil {
			return nil, err
		}

		metadata := Metadata{Data: data}
		custom := Custom{InToto: metadata}
		filename = gun + "/in-toto-pubkeys/" + filename
		publicKeys[filename] = custom
	}

	return publicKeys, nil
}

// GetRootLayout reads root layout off disk into a RootLayout struct
func GetRootLayout(gun string, filename string, publicKeys PublicKeys) (RootLayout, error) {
	var rootLayout RootLayout

	if !strings.HasSuffix(filename, ".layout") {
		return rootLayout, fmt.Errorf("%s does not have a .layout suffix", filename)
	}

	data, err := readFile(filename)
	if err != nil {
		return rootLayout, err
	}

	filenames := make([]string, len(publicKeys))
	for filename := range publicKeys {
		filenames = append(filenames, filename)
	}

	metadata := Metadata{Data: data, PublicKeys: filenames}
	custom := Custom{InToto: metadata}
	filename = gun + "/in-toto-metadata/" + filename
	rootLayout.Filename = filename
	rootLayout.Custom = custom
	return rootLayout, nil
}

// GetLinks reads link metadata off disk into a Links struct
func GetLinks(gun string, dir string) (Links, error) {
	fileinfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read links directory %v: %v", dir, err)
	}

	// NOTE: We assume that the base directory used to hold the links are unique
	// (e.g., identified by the digest of the first link metadata file corresponding to the first step).
	// This is so that different versions of links corresponding to the same root layout can be safely isolated from each other.
	// TODO: If they are not unique, should we raise an error later when adding them to the TUF targets metadata?
	digest := filepath.Base(dir)
	links := make(Links)
	for _, fileinfo := range fileinfos {
		filename := fileinfo.Name()
		if !strings.Contains(filename, ".link") {
			return nil, fmt.Errorf("%s does not have a .link suffix", filename)
		}

		data, err := readFile(filepath.Join(dir, filename))
		if err != nil {
			return nil, err
		}

		metadata := Metadata{Data: data}
		custom := Custom{InToto: metadata}
		filename = gun + "/in-toto-metadata/" + digest + "/" + filename
		links[filename] = custom
	}

	return links, nil
}

// GetBundleCustom returns a Custom struct pointing to a list of root layout and links
func GetBundleCustom(rootLayout RootLayout, links Links) Custom {
	filenames := make([]string, len(links)+1)
	filenames = append(filenames, rootLayout.Filename)
	for filename := range links {
		filenames = append(filenames, filename)
	}
	metadata := Metadata{Links: filenames}
	return Custom{InToto: metadata}
}

// WriteMetadataFiles writes the content of a metadata object into files in a directory
// TODO
func WriteMetadataFiles(m *Metadata, dir string) error {
	return fmt.Errorf("not implemented")
}

func readFile(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
