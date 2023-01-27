package types

var ActorTypes = []ActorType{
	ActorType_ACTOR_TYPE_APP,
	ActorType_ACTOR_TYPE_SERVICENODE,
	ActorType_ACTOR_TYPE_FISH,
	ActorType_ACTOR_TYPE_VAL,
}

func (a ActorType) GetName() string {
	return ActorType_name[int32(a)]
}
