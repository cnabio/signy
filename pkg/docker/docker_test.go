package docker

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto"

func TestRun(t *testing.T) {
	// NOTE: Tag will be empty since we cannot inject build-time variables during testing.
	// Therefore, we shall use the "latest" tag.
	err := Run(VerificationImage+"latest", testDir, log.InfoLevel.String())
	assert.NoError(t, err)
}
