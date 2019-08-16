package intoto

import (
	"fmt"

	"github.com/in-toto/in-toto-golang/in_toto"
)

// Verify performs the in-toto validation steps
func Verify(layout, linkDir string, layoutKeyPaths ...string) error {
	layoutKeys := make(map[string]in_toto.Key)
	for _, p := range layoutKeyPaths {
		var k in_toto.Key
		if err := k.LoadPublicKey(p); err != nil {
			return fmt.Errorf("cannot load layout public key %v: %v", p, err)
		}
		layoutKeys[k.KeyId] = k
	}

	var mb in_toto.Metablock
	if err := mb.Load(layout); err != nil {
		return fmt.Errorf("cannot load layout from file file %v: %v", layout, err)
	}

	if err := ValidateLayout(mb.Signed.(in_toto.Layout)); err != nil {
		return fmt.Errorf("invalid metatada found: %v", err)
	}

	if _, err := in_toto.InTotoVerifyWithDirectory(mb, layoutKeys, linkDir, linkDir, "", make(map[string]string)); err != nil {
		return fmt.Errorf("failed verification: %v", err)
	}

	fmt.Printf("Verification succeeded.\n")
	return nil
}
