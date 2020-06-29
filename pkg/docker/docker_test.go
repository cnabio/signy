package docker

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testDir = "../../testdata/intoto/minimal"

func TestRun(t *testing.T) {
	// NOTE: Tag will be empty since we cannot inject build-time variables during testing.
	// Therefore, we shall use the "latest" tag.
	err := Run(VerificationImage+"latest", testDir, log.InfoLevel.String())
	assert.NoError(t, err)
}

func TestParseReference(t *testing.T) {
	tests := []struct {
		input string
		gun   string
	}{
		{
			input: "localhost:5000/local-test-simple:v1",
			gun:   "localhost:5000/local-test-simple",
		},
		{
			input: "localhost:5000/multi-path/some/bundle:v1",
			gun:   "localhost:5000/multi-path/some/bundle",
		},
		{
			input: "dockerhubusername/bundle:v3",
			gun:   "docker.io/dockerhubusername/bundle",
		},
		{
			input: "mycnabregistry.azurecr.io/org/sub-org/bundle:latest",
			gun:   "mycnabregistry.azurecr.io/org/sub-org/bundle",
		},
	}

	is := assert.New(t)
	for _, test := range tests {
		gun, err := GetGUN(test.input)
		is.NoError(err)
		is.Equal(test.gun, gun)
	}
}
