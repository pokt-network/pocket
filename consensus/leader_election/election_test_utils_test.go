package leader_election

import (
	"testing"

	"github.com/pokt-network/pocket/consensus/leader_election/vrf"
	// types2 "github.com/pokt-network/pocket/shared/pkg/types"

	"github.com/stretchr/testify/require"
)

type TestValidatorConfigs struct {
	NodeId uint
	UPokt  uint
}

// TODO: Should there be a global value like this?
type ValidatorWithPrivateKeys struct {
	validator *types2.Validator
	privKey   *types2.PrivateKey
	secretKey *vrf.SecretKey
}

type ValMap map[types2.NodeId]*ValidatorWithPrivateKeys

func prepareTestValidators(t *testing.T, testValidatorConfigs []*TestValidatorConfigs) (valMap ValMap, totalStakedAmount uint64) {
	valMap = make(ValMap)
	for _, cfg := range testValidatorConfigs {
		privKey := types2.GeneratePrivateKey(uint32(cfg.NodeId))

		sk, vk, err := vrf.GenerateVRFKeys(nil)
		require.NoError(t, err)

		nodeId := types2.NodeId(cfg.NodeId)
		uPokt := uint64(cfg.UPokt)

		valMap[nodeId] = &ValidatorWithPrivateKeys{
			validator: &types2.Validator{
				NodeId:             nodeId,
				PublicKey:          privKey.Public(),
				UPokt:              uPokt,
				VRFVerificationKey: *vk,
			},
			privKey:   &privKey,
			secretKey: sk,
		}
		totalStakedAmount += uPokt
	}
	return
}
