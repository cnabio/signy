package cnab

import (
	"fmt"
	"os"
	"strings"

	containerdRemotes "github.com/containerd/containerd/remotes"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cnab-to-oci/remotes"
)

func createResolver(insecureRegistries []string) containerdRemotes.Resolver {
	return remotes.CreateResolver(config.LoadDefaultConfigFile(os.Stderr), insecureRegistries...)
}

func displayEvent(ev remotes.FixupEvent) {
	switch ev.EventType {
	case remotes.FixupEventTypeCopyImageStart:
		fmt.Fprintf(os.Stderr, "Starting to copy image %s...\n", ev.SourceImage)
	case remotes.FixupEventTypeCopyImageEnd:
		if ev.Error != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy image %s: %s\n", ev.SourceImage, ev.Error)
		} else {
			fmt.Fprintf(os.Stderr, "Completed image %s copy\n", ev.SourceImage)
		}
	}
}

// SplitTargetRef takes a target reference and returns a GUN and tag
// TODO - Radu M
//
// Should we error if the tag is not present, instead of returning latest?
func SplitTargetRef(ref string) (string, string) {
	parts := strings.Split(ref, ":")
	if len(parts) == 1 {
		parts = append(parts, "latest")
	}

	return parts[0], parts[1]
}
