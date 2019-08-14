package cnab

import (
	"fmt"
	"os"

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
