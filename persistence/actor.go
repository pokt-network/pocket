package persistence

import (
	"github.com/pokt-network/pocket/persistence/types"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

// TODO (#399): All of the functions below following a structure similar to `GetAll<Actor>`
// can easily be refactored and condensed into a single function using a generic type or a common
// interface.
func (p *PostgresContext) GetAllApps(height int64) (apps []*coreTypes.Actor, err error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, types.ApplicationActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*coreTypes.Actor
	for rows.Next() {
		var actor *coreTypes.Actor
		actor, height, err = p.getActorFromRow(types.ApplicationActor.GetActorType(), rows)
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

func (p *PostgresContext) GetAllValidators(height int64) (vals []*coreTypes.Actor, err error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, types.ValidatorActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*coreTypes.Actor
	for rows.Next() {
		var actor *coreTypes.Actor
		actor, height, err = p.getActorFromRow(types.ValidatorActor.GetActorType(), rows)
		if err != nil {
			return
		}
		actor.ActorType = types.ValidatorActor.GetActorType()
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

func (p *PostgresContext) GetAllServicers(height int64) (sn []*coreTypes.Actor, err error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, types.ServicerActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*coreTypes.Actor
	for rows.Next() {
		var actor *coreTypes.Actor
		actor, height, err = p.getActorFromRow(types.ServicerActor.GetActorType(), rows)
		if err != nil {
			return
		}
		actors = append(actors, actor)
	}
	rows.Close()
	for _, actor := range actors {
		actor, err = p.getChainsForActor(ctx, tx, types.ServicerActor, actor, height)
		if err != nil {
			return
		}
		sn = append(sn, actor)
	}
	return
}

func (p *PostgresContext) GetAllFishermen(height int64) (f []*coreTypes.Actor, err error) {
	ctx, tx := p.getCtxAndTx()
	rows, err := tx.Query(ctx, types.FishermanActor.GetAllQuery(height))
	if err != nil {
		return nil, err
	}
	var actors []*coreTypes.Actor
	for rows.Next() {
		var actor *coreTypes.Actor
		actor, height, err = p.getActorFromRow(types.FishermanActor.GetActorType(), rows)
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
func (p *PostgresContext) GetAllStakedActors(height int64) (allActors []*coreTypes.Actor, err error) {
	type actorGetter func(height int64) ([]*coreTypes.Actor, error)
	actorGetters := []actorGetter{p.GetAllValidators, p.GetAllServicers, p.GetAllFishermen, p.GetAllApps}
	for _, actorGetter := range actorGetters {
		var actors []*coreTypes.Actor
		actors, err = actorGetter(height)
		if err != nil {
			return
		}
		allActors = append(allActors, actors...)
	}
	return
}
