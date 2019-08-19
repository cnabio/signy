package intoto

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto"

func TestVerify(t *testing.T) {
	layout := filepath.Join(testDir, "demo.layout.template")
	key := filepath.Join(testDir, "alice.pub")
	links := testDir

	err := Verify(layout, links, key)
	assert.NoError(t, err)

	// the verification step geneates a file called untar.link
	os.Remove("untar.link")
}

func TestValidate(t *testing.T) {
	layoutPath := filepath.Join(testDir, "demo.layout.template")

	l, err := getLayout(layoutPath)
	assert.NoError(t, err)

	err = ValidateLayout(*l)
	assert.NoError(t, err)
}

func TestValidateMalformed(t *testing.T) {
	layoutPath := filepath.Join(testDir, "malformed.template")

	// we can load the file and unmarshal a layout
	l, err := getLayout(layoutPath)
	assert.NoError(t, err)

	// malformed.template is missing signatures, so validation should fail
	err = ValidateLayout(*l)
	assert.Error(t, err)
}
