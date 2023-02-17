package main

import (
	cryptoPocket "github.com/pokt-network/pocket/shared/crypto"
	v1 "k8s.io/api/core/v1"
)

func isValidator(service *v1.Service) bool {
	return service.Labels["v1-purpose"] == "validator"
}

func extractValidatorId(validatorName string) string {
	if len(validatorName) >= 3 {
		return validatorName[len(validatorName)-3:]
	}
	return validatorName
}

func getPrivateKey(validatorKeysMap map[string]string, validatorId string) cryptoPocket.PrivateKey {
	privHexString := validatorKeysMap[validatorId]
	keyPair, err := cryptoPocket.CreateNewKeyFromString(privHexString, "", "")
	if err != nil {
		panic(err)
	}

	privateKey, err := keyPair.Unarmour("")
	if err != nil {
		logger.Err(err).Msg("Error unarmouring private key")
	}
	return privateKey
}
