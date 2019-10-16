package cnab

import (
	"os"
	"strings"

	containerdRemotes "github.com/containerd/containerd/remotes"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cnab-to-oci/remotes"
	log "github.com/sirupsen/logrus"
)

func createResolver(insecureRegistries []string) containerdRemotes.Resolver {
	return remotes.CreateResolver(config.LoadDefaultConfigFile(os.Stderr), insecureRegistries...)
}

func displayEvent(ev remotes.FixupEvent) {
	switch ev.EventType {
	case remotes.FixupEventTypeCopyImageStart:
		log.Infof("Starting to copy image %s", ev.SourceImage)
	case remotes.FixupEventTypeCopyImageEnd:
		if ev.Error != nil {
			log.Infof("Failed to copy image %s: %s", ev.SourceImage, ev.Error)
		} else {
			log.Infof("Completed image %s copy", ev.SourceImage)
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
