package canonicaljson

import (
	"encoding/json"

	canonicaljson "github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	rawmessage "github.com/docker/go/canonical/json" // We are only using this library for the RawMessage type that TUF uses, the actual marshaling is done by the webpki.org/jsoncanonicalizer library
)

// define a type alias for raw message in this package
// so that outside of this package we only import a single canonical json library.
type RawMessage = rawmessage.RawMessage

// Marshal returns the canonical json encoded value of v.
func Marshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return canonicaljson.Transform(b)
}

func MarshalToRawMessage(v interface{}) (rawmessage.RawMessage, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}
