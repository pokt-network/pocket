package types

func GetActorName(actor ActorType) string {
	return ActorType_name[int32(actor)]
}
