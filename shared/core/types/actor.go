package types

func (a ActorType) GetName() string {
	return ActorType_name[int32(a)]
}

func (a StakeStatus) GetName() string {
	return ActorType_name[int32(a)]
}
