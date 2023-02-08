package types

// TODO(Optional): Evaluate a way to define an interface for the actors and remove all the `switch actorType`
// statements throughout the codebase.

var ActorTypes = []ActorType{
	ActorType_ACTOR_TYPE_APP,
	ActorType_ACTOR_TYPE_FISH,
	ActorType_ACTOR_TYPE_SERVICENODE,
	ActorType_ACTOR_TYPE_VAL,
}

func (a ActorType) GetName() string {
	return ActorType_name[int32(a)]
}
