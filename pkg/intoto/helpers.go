package intoto

import (
	"fmt"
	"regexp"

	"github.com/in-toto/in-toto-golang/in_toto"
)

// validateRSAPubKey checks if a passed key is a valid RSA public key.
func validateRSAPubKey(key in_toto.Key) error {
	if key.KeyType != "rsa" {
		return fmt.Errorf("invalid KeyType for key '%s': should be 'rsa', got"+
			" '%s'", key.KeyId, key.KeyType)
	}
	if key.Scheme != "rsassa-pss-sha256" {
		return fmt.Errorf("invalid scheme for key '%s': should be "+
			"'rsassa-pss-sha256', got: '%s'", key.KeyId, key.Scheme)
	}
	if err := validatePubKey(key); err != nil {
		return err
	}
	return nil
}

// validatePubKey is a general function to validate if a key is a valid public key.
func validatePubKey(key in_toto.Key) error {
	if err := validateHexString(key.KeyId); err != nil {
		return fmt.Errorf("keyid: %s", err.Error())
	}
	if key.KeyVal.Private != "" {
		return fmt.Errorf("in key '%s': private key found", key.KeyId)
	}
	if key.KeyVal.Public == "" {
		return fmt.Errorf("in key '%s': public key cannot be empty", key.KeyId)
	}
	return nil
}

// validateHexString is used to validate that a string passed to it contains
// only valid hexadecimal characters.
func validateHexString(str string) error {
	formatCheck, _ := regexp.MatchString("^[a-fA-F0-9]+$", str)
	if !formatCheck {
		return fmt.Errorf("'%s' is not a valid hex string", str)
	}
	return nil
}

// validateStep ensures that a passed step is valid and matches the
// necessary format of an step.
func validateStep(step in_toto.Step) error {
	if err := validateSupplyChainItem(step.SupplyChainItem); err != nil {
		return fmt.Errorf("step %s", err.Error())
	}
	if step.Type != "step" {
		return fmt.Errorf("invalid Type value for step '%s': should be 'step'",
			step.SupplyChainItem.Name)
	}
	for _, keyID := range step.PubKeys {
		if err := validateHexString(keyID); err != nil {
			return fmt.Errorf("in step '%s', keyid: %s",
				step.SupplyChainItem.Name, err.Error())
		}
	}
	return nil
}

// validateSupplyChainItem is used to validate the common elements found in both
// steps and inspections. Here, the function primarily ensures that the name of
// a supply chain item isn't empty.
func validateSupplyChainItem(item in_toto.SupplyChainItem) error {
	if item.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if err := validateSliceOfArtifactRules(item.ExpectedMaterials); err != nil {
		return fmt.Errorf("invalid material rule: %s", err)
	}
	if err := validateSliceOfArtifactRules(item.ExpectedProducts); err != nil {
		return fmt.Errorf("invalid product rule: %s", err)
	}
	return nil
}

// validateArtifactRule calls UnpackRule to validate that the passed rule conforms
// with any of the available rule formats.
func validateArtifactRule(rule []string) error {
	if _, err := in_toto.UnpackRule(rule); err != nil {
		return err
	}
	return nil
}

// validateSliceOfArtifactRules iterates over passed rules to validate them.
func validateSliceOfArtifactRules(rules [][]string) error {
	for _, rule := range rules {
		if err := validateArtifactRule(rule); err != nil {
			return err
		}
	}
	return nil
}

// GetLayout returns an In-Toto layout given a file path
func GetLayout(layout string) (*in_toto.Layout, error) {
	var mb in_toto.Metablock
	if err := mb.Load(layout); err != nil {
		return nil, fmt.Errorf("cannot load layout from file file %v: %v", layout, err)
	}
	l := mb.Signed.(in_toto.Layout)
	return &l, nil
}
