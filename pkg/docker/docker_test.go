package docker

import (
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto"

func TestRun(t *testing.T) {
	image := "radumatei/in-toto-container:v1"
	layout := filepath.Join(testDir, "demo.layout.template")
	key := filepath.Join(testDir, "alice.pub")
	links := testDir
	targets := []string{filepath.Join(testDir, "foo.tar.gz")}

	err := Run(image, layout, key, links, log.InfoLevel.String(), targets)
	assert.NoError(t, err)
}
