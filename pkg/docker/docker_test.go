package docker

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto"

func TestRun(t *testing.T) {
	err := Run(VerificationImage, testDir, log.InfoLevel.String())
	assert.NoError(t, err)
}
