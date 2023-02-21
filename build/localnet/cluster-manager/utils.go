package main

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	v1 "k8s.io/api/core/v1"
)

func isValidator(service *v1.Service) bool {
	return service.Labels["v1-purpose"] == "validator"
}

// extractValidatorId extracts the validator id from the validator name (e.g. v1-validator001 -> 001)
//
// it follows the pattern defined in the v1-validator template (/build/localnet/templates/v1-validator-template.yaml.tpl)
func extractValidatorId(validatorName string) string {
	if len(validatorName) >= 3 {
		return validatorName[len(validatorName)-3:]
	}
	return validatorName
}

// TODO: Create a type for `validatorKeyMap` and document what the expected key-value types contain
func getPrivateKey(validatorKeysMap map[string]string, validatorId string) cryptoPocket.PrivateKey {
	privHexString := validatorKeysMap[validatorId]
	keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
	if err != nil {
		panic(err)
	}

	privateKey, err := keyPair.Unarmour("") // empty passphrase
	if err != nil {
		logger.Err(err).Msg("Error unarmouring private key")
	}
	return privateKey
}
