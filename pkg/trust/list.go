package trust

import (
	"encoding/hex"
	"fmt"

	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

// PrintTargets prints all the targets for a specific GUN from a trust server
func PrintTargets(gun, trustServer, tlscacert, trustDir string) error {
	if err := ensureTrustDir(trustDir); err != nil {
		return fmt.Errorf("cannot ensure trust directory: %v", err)
	}

	transport, err := makeTransport(trustServer, gun, tlscacert)
	if err != nil {
		return fmt.Errorf("cannot make transport: %v", err)
	}

	repo, err := client.NewFileCachedRepository(
		trustDir,
		data.GUN(gun),
		trustServer,
		transport,
		nil,
		trustpinning.TrustPinConfig{},
	)
	if err != nil {
		return fmt.Errorf("cannot create new file cached repository: %v", err)
	}

	targets, err := repo.ListTargets()
	if err != nil {
		return fmt.Errorf("cannot list targets:%v", err)
	}

	for _, tgt := range targets {
		fmt.Printf("%s\t%s\n", tgt.Name, hex.EncodeToString(tgt.Hashes["sha256"]))
	}

	return nil
}
