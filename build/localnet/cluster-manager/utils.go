package main

import (
	"regexp"

	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	v1 "k8s.io/api/core/v1"
)

func isValidator(service *v1.Service) bool {
	return service.Labels["pokt.network/purpose"] == "validator"
}

// extractValidatorId extracts the validator id from the validator name (e.g. validator-001-pocket -> 001)
//
// it follows the pattern defined in the pocket helm chart.
func extractValidatorId(validatorName string) string {
	if len(validatorName) >= 3 {
		re := regexp.MustCompile(`validator-(\d+)-pocket`)
		return re.FindStringSubmatch(validatorName)[1]
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

// shouldSkipAutoStaking returns true if the validatorId is in the autoStakeSkipStakeForValidatorIds list
func shouldSkipAutoStaking(validatorId string) bool {
	for _, id := range autoStakeSkipStakeForValidatorIds {
		if id == validatorId {
			return true
		}
	}
	return false
}
