package leader_election

import (
	"fmt"
	"testing"

	"github.com/pokt-network/pocket/consensus/leader_election/vrf"
	types_consensus "github.com/pokt-network/pocket/consensus/types"
	"github.com/pokt-network/pocket/shared/types"
)

type TestValidatorConfigs struct {
	NodeId uint
	UPokt  uint
}

// TODO: Should there be a global value like this?
type ValidatorWithPrivateKeys struct {
	validator *types.Validator
	// privKey   *crypto.PrivateKey
	secretKey *vrf.SecretKey
}

type ValMap map[types_consensus.NodeId]*ValidatorWithPrivateKeys

func prepareTestValidators(t *testing.T, testValidatorConfigs []*TestValidatorConfigs) (valMap ValMap, totalStakedAmount uint64) {
	valMap = make(ValMap)
	for _, cfg := range testValidatorConfigs {
		fmt.Println(cfg)
		// privKey := pcrypto.NewPri() GeneratePrivateKey(uint32(cfg.NodeId))

		// sk, _, err := vrf.GenerateVRFKeys(nil)
		// require.NoError(t, err)

		// nodeId := types_consensus.NodeId(cfg.NodeId)
		// uPokt := uint64(cfg.UPokt)

		// valMap[nodeId] = &ValidatorWithPrivateKeys{
		// 	validator: &types.Validator{
		// 		// NodeId:             nodeId,
		// 		PublicKey: privKey.Public(),
		// 		UPokt:     uPokt,
		// 		// VRFVerificationKey: *vk,
		// 	},
		// 	privKey:   &privKey,
		// 	secretKey: sk,
		// }
		// totalStakedAmount += uPokt
	}
	return
}
