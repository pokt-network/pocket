package types

func (a ActorType) GetName() string {
	return ActorType_name[int32(a)]
}

func (a ActorType) GetNameShort() string {
	return ActorType_name_short[int32(a)]
}
