package intoto

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto"

func TestVerify(t *testing.T) {
	err := verifyOnOS(testDir)
	assert.NoError(t, err)

	// the verification step generates a file called untar.link
	os.Remove("untar.link")
}

func TestValidate(t *testing.T) {
	layoutPath := filepath.Join(testDir, "root.layout")

	l, err := getLayout(layoutPath)
	assert.NoError(t, err)

	err = ValidateLayout(*l)

	// the validation step generates a directory
	os.RemoveAll(testDir + "/demo-project")

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
