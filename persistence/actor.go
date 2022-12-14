package persistence

import (
	"github.com/pokt-network/pocket/persistence/types"
	"github.com/pokt-network/pocket/shared/modules"
)

// TODO(pocket/issues/149): All of the functions below following a structure similar to `GetAll<Actor>`
//	can easily be refactored and condensed into a single function using a generic type or a common
//  interface.
func (p PostgresContext) GetAllApps(height int64) (apps []modules.Actor, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, types.ApplicationActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*types.Actor
	for rows.Next() {
		var actor *types.Actor
		actor, height, err = p.getActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actorWithChains, err := p.getChainsForActor(ctx, tx, types.ApplicationActor, actor, height)
		if err != nil {
			return nil, err
		}
		apps = append(apps, actorWithChains)
	}
	return
}

func (p PostgresContext) GetAllValidators(height int64) (vals []modules.Actor, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, types.ValidatorActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*types.Actor
	for rows.Next() {
		var actor *types.Actor
		actor, height, err = p.getActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.getChainsForActor(ctx, tx, types.ValidatorActor, actor, height)
		if err != nil {
			return
		}
		vals = append(vals, actor)
	}
	return
}

func (p PostgresContext) GetAllServiceNodes(height int64) (sn []modules.Actor, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, types.ServiceNodeActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*types.Actor
	for rows.Next() {
		var actor *types.Actor
		actor, height, err = p.getActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.getChainsForActor(ctx, tx, types.ServiceNodeActor, actor, height)
		if err != nil {
			return
		}
		sn = append(sn, actor)
	}
	return
}

func (p PostgresContext) GetAllFishermen(height int64) (f []modules.Actor, err error) {
	ctx, tx, err := p.getCtxAndTx()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, types.FishermanActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*types.Actor
	for rows.Next() {
		var actor *types.Actor
		actor, height, err = p.getActorFromRow(rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.getChainsForActor(ctx, tx, types.FishermanActor, actor, height)
		if err != nil {
			return
		}
		f = append(f, actor)
	}
	return
}

// IMPROVE: This is a proof of concept. Ideally we should have a single query that returns all actors.
func (p PostgresContext) GetAllStakedActors(height int64) (allActors []modules.Actor, err error) {
	type actorGetter func(height int64) ([]modules.Actor, error)
	actorGetters := []actorGetter{p.GetAllValidators, p.GetAllServiceNodes, p.GetAllFishermen, p.GetAllApps}
	for _, actorGetter := range actorGetters {
		actors, err := actorGetter(height)
		if err != nil {
			allActors = append(allActors, actors...)
		}
	}
	return
}
