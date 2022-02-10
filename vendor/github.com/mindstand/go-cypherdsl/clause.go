package go_cypherdsl

type Cypher interface {
	Match
	OptionalMatch
	Create
	Where
	Merge
	Return
	Delete
	Set
	Remove
	OrderBy
	Limit
	Skip
	With
	Unwind
	Union
	CustomCypher
	QueryCompleter
}

//complete
type Match interface {
	Match(p *PathBuilder) Cypher
}

type OptionalMatch interface {
	OptionalMatch(p *PathBuilder) Cypher
}

//complete
type Create interface {
	Create(CreateQuery, error) Cypher
}

//complete
type Where interface {
	Where(cb ConditionOperator) Cypher
}

//complete
type Merge interface {
	Merge(mergeConf *MergeConfig) Cypher
}

//complete
type Return interface {
	Return(distinct bool, parts ...ReturnPart) Cypher
}

//complete
type Delete interface {
	Delete(detach bool, params ...string) Cypher
}

//complete
type Set interface {
	Set(sets ...SetConfig) Cypher
}

type Remove interface {
	Remove(removes ...RemoveConfig) Cypher
}

type OrderBy interface {
	OrderBy(orderBys ...OrderByConfig) Cypher
}

type Limit interface {
	Limit(num int) Cypher
}

type Skip interface {
	Skip(num int) Cypher
}

type With interface {
	With(conf *WithConfig) Cypher
}

type Unwind interface {
	Unwind(unwind *UnwindConfig) Cypher
}

type Union interface {
	Union(all bool) Cypher
}

type CustomCypher interface {
	Cypher(q string) Cypher
}

type QueryCompleter interface {
	ToCypher() (string, error)
}
