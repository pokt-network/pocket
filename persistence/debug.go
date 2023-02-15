package persistence

import (
	"context"
	"crypto/sha256"
	"log"
	"runtime/debug"

	"github.com/celestiaorg/smt"
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/codec"
	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/messaging"
)

// A list of functions to clear data from the DB not associated with protocol actors
var nonActorClearFunctions = []func() string{
	types.Account.ClearAllAccounts,
	types.Pool.ClearAllAccounts,
	types.ClearAllGovParamsQuery,
	types.ClearAllGovFlagsQuery,
	types.ClearAllBlocksQuery,
}

func (m *persistenceModule) HandleDebugMessage(debugMessage *messaging.DebugMessage) error {
	switch debugMessage.Action {
	// Prints the block with the largest height from the KV store
	case messaging.DebugMessageAction_DEBUG_SHOW_LATEST_BLOCK_IN_STORE:
		return m.showLatestBlockInStore(debugMessage)

	// Prints the block with the largest height from the KV store
	case messaging.DebugMessageAction_DEBUG_EXPORT_TO_NEO:
		return m.exportToNeo(debugMessage)

	// Clears all the state (SQL DB, KV Stores, Trees, etc) to nothing
	case messaging.DebugMessageAction_DEBUG_PERSISTENCE_CLEAR_STATE:
		return m.clearAllState(debugMessage)

	// Resets all the state (SQL DB, KV Stores, Trees, etc) to the tate specified in the genesis file provided
	case messaging.DebugMessageAction_DEBUG_PERSISTENCE_RESET_TO_GENESIS:
		if err := m.clearAllState(debugMessage); err != nil {
			return err
		}
		g := m.genesisState
		m.populateGenesisState(g) // fatal if there's an error

	// Not handled yet
	default:
		m.logger.Debug().Str("message", debugMessage.Message.String()).Msg("Debug message not handled by persistence module")
	}

	return nil
}

func dropAllNeo() {
	driver, err := neo4j.NewDriver("bolt://neo4j:7687", neo4j.BasicAuth("root", "", ""), func(c *neo4j.Config) { c.Encrypted = false })
	if err != nil {
		log.Panic(err)
	}
	defer driver.Close()

	session, err := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	// MATCH (n:Actor) RETURN n LIMIT 100
	res, err := session.Run("MATCH (n:NeoPool) DETACH DELETE n", map[string]interface{}{})
	if err != nil {
		log.Panic(err)
	}
	log.Println(res)

	res, err = session.Run("MATCH (n:NeoActor) DETACH DELETE n", map[string]interface{}{})
	if err != nil {
		log.Panic(err)
	}
	log.Println(res)

	res, err = session.Run("MATCH (n:NeoAccount) DETACH DELETE n", map[string]interface{}{})
	if err != nil {
		log.Panic(err)
	}
	log.Println(res)
}

// See https://github.com/mindstand/gogm?ref=golangrepo.com as a reference
func (m *persistenceModule) exportToNeo(_ *messaging.DebugMessage) error {
	config := gogm.Config{
		Host:     "neo4j",
		Port:     7687,
		Protocol: "bolt", // {neo4j neo4j+s, neo4j+ssc, bolt, bolt+s and bolt+ssc}
		Username: "neo4j",
		// Password:           "",
		PoolSize: 50,
		// IndexStrategy:      gogm.VALIDATE_INDEX, // {VALIDATE_INDEX, ASSERT_INDEX, IGNORE_INDEX}
		IndexStrategy:      gogm.IGNORE_INDEX, // {VALIDATE_INDEX, ASSERT_INDEX, IGNORE_INDEX}
		TargetDbs:          nil,
		Logger:             gogm.GetDefaultLogger(),
		LogLevel:           "DEBUG",
		EnableDriverLogs:   false,
		EnableLogParams:    false,
		OpentracingEnabled: false,
	}

	_gogm, err := gogm.New(&config, gogm.DefaultPrimaryKeyStrategy, &coreTypes.NeoActor{}, &coreTypes.NeoAccount{}, &coreTypes.NeoPool{})
	if err != nil {
		return err
	}

	session, err := _gogm.NewSessionV2(gogm.SessionConfig{AccessMode: gogm.AccessModeWrite})
	if err != nil {
		return err
	}
	defer session.Close()

	readCtx, err := m.NewReadContext(-1)
	if err != nil {
		return err
	}

	latestHeight, err := readCtx.GetLatestBlockHeight()
	if err != nil {
		return err
	}

	allActors, err := readCtx.GetAllStakedActors(int64(latestHeight))
	if err != nil {
		return err
	}

	allPools, err := readCtx.GetAllPools(int64(latestHeight))
	if err != nil {
		return err
	}

	allAccounts, err := readCtx.GetAllAccounts(int64(latestHeight))
	if err != nil {
		return err
	}

	// Drop all existing nodes in the neo4j DB
	dropAllNeo()

	sessionCtx := context.Background()
	for _, actor := range allActors {
		neoActor := &coreTypes.NeoActor{
			ActorType: actor.ActorType.GetNameShort(),
			Address:   actor.Address,
			// PublicKey: actor.PublicKey,
			// Chains:          actor.Chains,
			// GenericParam:    actor.GenericParam,
			StakedAmount: actor.StakedAmount,
			// PausedHeight:    actor.PausedHeight,
			// UnstakingHeight: actor.UnstakingHeight,
			// Output:          actor.Output,
		}

		if err = session.Save(sessionCtx, neoActor); err != nil {
			return err
		}
	}

	for _, pool := range allPools {
		neoPool := &coreTypes.NeoPool{
			Name:   pool.Address,
			Amount: pool.Amount,
		}
		if err = session.Save(sessionCtx, neoPool); err != nil {
			return err
		}
	}

	for _, account := range allAccounts {
		neoAccount := &coreTypes.NeoAccount{
			Address: account.Address,
			Amount:  account.Amount,
		}
		if err = session.Save(sessionCtx, neoAccount); err != nil {
			return err
		}
	}

	return nil
}

func (m *persistenceModule) showLatestBlockInStore(_ *messaging.DebugMessage) error {
	// TODO: Add an iterator to the `kvstore` and use that instead
	height := m.GetBus().GetConsensusModule().CurrentHeight() - 1
	blockBytes, err := m.GetBlockStore().Get(converters.HeightToBytes(height))
	if err != nil {
		m.logger.Error().Err(err).Uint64("height", height).Msg("Error getting block from block store")
		return err
	}

	block := &coreTypes.Block{}
	if err := codec.GetCodec().Unmarshal(blockBytes, block); err != nil {
		m.logger.Error().Err(err).Uint64("height", height).Msg("Error decoding block from block store")
		return err
	}

	m.logger.Info().Uint64("height", height).Str("block", block.String()).Msg("Block from block store")
}

// TECHDEBT: Make sure this is atomic
func (m *persistenceModule) clearAllState(_ *messaging.DebugMessage) error {
	ctx, err := m.NewRWContext(-1)
	if err != nil {
		return err
	}
	postgresCtx := ctx.(*PostgresContext)

	// Clear the SQL DB
	if err := postgresCtx.clearAllSQLState(); err != nil {
		return err
	}

	// Release the SQL context
	if err := m.ReleaseWriteContext(); err != nil {
		return err
	}

	// Clear the BlockStore
	if err := m.blockStore.ClearAll(); err != nil {
		return err
	}

	// Clear all the Trees
	if err := postgresCtx.clearAllTreeState(); err != nil {
		return err
	}

	m.logger.Info().Msg("Cleared all the state")
	// reclaming memory manually because the above calls deallocate and reallocate a lot of memory
	debug.FreeOSMemory()
	return nil
}

func (p *PostgresContext) clearAllSQLState() error {
	ctx, clearTx := p.getCtxAndTx()

	for _, actor := range protocolActorSchemas {
		if _, err := clearTx.Exec(ctx, actor.ClearAllQuery()); err != nil {
			return err
		}
		if actor.GetChainsTableName() != "" {
			if _, err := clearTx.Exec(ctx, actor.ClearAllChainsQuery()); err != nil {
				return err
			}
		}
	}

	for _, clearFn := range nonActorClearFunctions {
		if _, err := clearTx.Exec(ctx, clearFn()); err != nil {
			return err
		}
	}

	if err := clearTx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *PostgresContext) clearAllTreeState() error {
	for treeType := merkleTree(0); treeType < numMerkleTrees; treeType++ {
		valueStore := p.stateTrees.valueStores[treeType]
		nodeStore := p.stateTrees.nodeStores[treeType]

		if err := valueStore.ClearAll(); err != nil {
			return err
		}
		if err := nodeStore.ClearAll(); err != nil {
			return err
		}

		// Needed in order to make sure the root is re-set correctly after clearing
		p.stateTrees.merkleTrees[treeType] = smt.NewSparseMerkleTree(valueStore, nodeStore, sha256.New())
	}

	return nil
}
