package types

// REFACTOR: Evaluate a way to define an interface for the actors which would enable to easily
//           remove all the `switch actorType` statements in the codebase.

var ActorTypes = []ActorType{
	ActorType_ACTOR_TYPE_APP,
	ActorType_ACTOR_TYPE_FISH,
	ActorType_ACTOR_TYPE_SERVICENODE,
	ActorType_ACTOR_TYPE_VAL,
}

func (a ActorType) GetName() string {
	return ActorType_name[int32(a)]
}
