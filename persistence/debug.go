package persistence

import (
	"crypto/sha256"
	"log"
	"os"
	"strings"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/debug"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (m *persistenceModule) HandleDebugMessage(debugMessage *debug.DebugMessage) error {
	switch debugMessage.Action {
	case debug.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		m.showLatestBlockInStore(debugMessage)
	case debug.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE:
		if err := m.clearState(debugMessage); err != nil {
			return err
		}
	case debug.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS:
		if err := m.clearState(debugMessage); err != nil {
			return err
		}
		g := m.genesisState.(*types.PersistenceGenesisState)
		m.populateGenesisState(g) // fatal if there's an error
	case debug.DebugMessageAction_DEBUG_PERSISTENCE_TREE_EXPORT:
		if err := m.exportTrees(debugMessage); err != nil {
			return err
		}
	default:
		log.Printf("Debug message not handled by persistence module: %s \n", debugMessage.Message)
	}
	return nil
}

func (m *persistenceModule) showLatestBlockInStore(_ *debug.DebugMessage) {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight()
	blockBytes, err := m.GetBlockStore().Get(heightToBytes(int64(height)))
	if err != nil {
		log.Printf("Error getting block %d from block store: %s \n", height, err)
		return
	}
	codec := codec.GetCodec()
	block := &types.Block{}
	codec.Unmarshal(blockBytes, block)

	log.Printf("Block at height %d: %+v \n", height, block)
}

// Everyone roles their own key-value export: https://www.reddit.com/r/golang/comments/bw08dt/is_there_any_offline_database_viewer_and_editor
// docker exec node4.consensus sh -c "cat /tmp/trees/val.json";
func (m *persistenceModule) exportTrees(_ *debug.DebugMessage) error {
	if err := os.Mkdir("/tmp/trees", os.ModePerm); err != nil {
		return err
	}

	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		smtValues := m.stateTrees.valueStores[treeType]
		_, values, err := smtValues.GetAll(nil, true)
		if err != nil {
			return err
		}

		var sb strings.Builder
		for i := 0; i < len(values); i++ {
			vProto := merkleTreeToProtoSchema[treeType]()
			// vProto := &types.Actor{}
			if err := proto.Unmarshal(values[i], vProto.(proto.Message)); err != nil {
				// if err := proto.Unmarshal(values[i], vProto); err != nil {
				return err
			}
			sb.WriteString(protojson.Format(vProto))
		}
		f, err := os.Create("/tmp/trees/" + merkleTreeToString[treeType] + ".json")
		if err != nil {
			return err
		}
		f.Write([]byte(sb.String()))
		f.Close()
	}
	log.Println("Trees exported to /tmp/trees/")
	return nil
}

// TODO: MAke sure this is atomic
func (m *persistenceModule) clearState(_ *debug.DebugMessage) error {
	context, err := m.NewRWContext(-1)
	if err != nil {
		return err
	}

	// Clear the SQL DB
	if err := context.(*PostgresContext).DebugClearAll(); err != nil {
		return err
	}

	if err := m.ReleaseWriteContext(); err != nil {
		return err
	}

	// Clear the KV Stores

	if err := m.blockStore.ClearAll(); err != nil {
		return err
	}

	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		valueStore := m.stateTrees.valueStores[treeType]
		nodeStore := m.stateTrees.nodeStores[treeType]

		if err := valueStore.ClearAll(); err != nil {
			return err
		}
		if err := nodeStore.ClearAll(); err != nil {
			return err
		}

		// Needed in order to make sure the root is re-set correctly after clearing
		m.stateTrees.merkleTrees[treeType] = smt.NewSparseMerkleTree(valueStore, nodeStore, sha256.New())
	}

	// if _, err := context.UpdateAppHash(); err != nil {
	// 	return err
	// }

	log.Println("Cleared all the state")

	return nil
}
