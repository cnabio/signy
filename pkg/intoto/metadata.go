package intoto

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	canonicaljson "github.com/docker/go/canonical/json"
)

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

	//FIXME: no need to actually write filename.
	err = ioutil.WriteFile(filepath.Join(abs, "root.layout"), m.Layout, ReadOnlyMask)
	if err != nil {
		return err
	}

	//FIXME: no need to actually write filenames.
	err = ioutil.WriteFile(filepath.Join(abs, "root.layout.pub"), m.Key, ReadOnlyMask)
	if err != nil {
		return err
	}

	for n, c := range m.Links {
		err = ioutil.WriteFile(filepath.Join(abs, n), c, ReadOnlyMask)
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
