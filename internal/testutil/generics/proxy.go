package generics_testutil

type ProxyFactory[T any] func(target T) (proxy T)
