package types

// REFACTOR: Evaluate a way to define an interface for the actors which would enable to easily
//
//	remove all the `switch actorType` statements in the codebase.
func (a ActorType) GetName() string {
	return ActorType_name[int32(a)]
}
