package leader_election

import (
	"testing"

	"pocket/consensus/pkg/consensus/leader_election/vrf"
	"pocket/consensus/pkg/types"

	"github.com/stretchr/testify/require"
)

type TestValidatorConfigs struct {
	NodeId uint
	UPokt  uint
}

// TODO: Should there be a global value like this?
type ValidatorWithPrivateKeys struct {
	validator *types.Validator
	privKey   *types.PrivateKey
	secretKey *vrf.SecretKey
}

type ValMap map[types.NodeId]*ValidatorWithPrivateKeys

func prepareTestValidators(t *testing.T, testValidatorConfigs []*TestValidatorConfigs) (valMap ValMap, totalVotingPower uint64) {
	valMap = make(ValMap)
	for _, cfg := range testValidatorConfigs {
		privKey := types.GeneratePrivateKey(uint32(cfg.NodeId))

		sk, vk, err := vrf.GenerateVRFKeys(nil)
		require.NoError(t, err)

		nodeId := types.NodeId(cfg.NodeId)
		uPokt := uint64(cfg.UPokt)

		valMap[nodeId] = &ValidatorWithPrivateKeys{
			validator: &types.Validator{
				NodeId:             nodeId,
				PublicKey:          privKey.Public(),
				UPokt:              uPokt,
				VRFVerificationKey: *vk,
			},
			privKey:   &privKey,
			secretKey: sk,
		}
		totalVotingPower += uPokt
	}
	return
}
