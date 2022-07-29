package types

func ActorName(actorType ActorType) string {
	return ActorType_name[int32(actorType)]
}
