package intoto

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto"

func TestVerify(t *testing.T) {
	err := verifyOnOS(path.Join(testDir, "minimal"))
	assert.NoError(t, err)

	// the verification step generates a file called untar.link
	os.Remove("untar.link")
}

func TestValidate(t *testing.T) {
	layoutPath := filepath.Join(testDir, "minimal", "root.layout")

	l, err := getLayout(layoutPath)
	assert.NoError(t, err)

	err = ValidateLayout(*l)
	assert.NoError(t, err)
}

func TestValidateMalformed(t *testing.T) {
	layoutPath := filepath.Join(testDir, "malformed.root.layout")

	// we can load the file and unmarshal a layout
	l, err := getLayout(layoutPath)
	assert.NoError(t, err)

	// malformed.template is missing signatures, so validation should fail
	err = ValidateLayout(*l)
	assert.Error(t, err)
}
