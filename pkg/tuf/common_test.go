package tuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		input      string
		repository string
		registry   string
		tag        string
	}{
		{
			input:      "localhost:5000/local-test-simple:v1",
			repository: "localhost:5000/local-test-simple",
			registry:   "localhost:5000",
			tag:        "v1",
		},
		{
			input:      "localhost:5000/multi-path/some/bundle:v1",
			repository: "localhost:5000/multi-path/some/bundle",
			registry:   "localhost:5000",
			tag:        "v1",
		},
		{
			input:      "dockerhubusername/bundle:v3",
			repository: "docker.io/dockerhubusername/bundle",
			registry:   "docker.io",
			tag:        "v3",
		},
		{
			input:      "mycnabregistry.azurecr.io/org/sub-org/bundle:latest",
			repository: "mycnabregistry.azurecr.io/org/sub-org/bundle",
			registry:   "mycnabregistry.azurecr.io",
			tag:        "latest",
		},
	}

	is := assert.New(t)
	for _, test := range tests {
		repo, tag, err := getRepoAndTag(test.input)
		is.NoError(err)
		is.Equal(test.repository, repo.Name.Name())
		is.Equal(test.tag, tag)
		is.Equal(test.registry, repo.Index.Name)
	}
}
