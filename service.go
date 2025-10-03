package ioc

type ScopedAdditional struct {
	Scope ScopeID
}

type SingletonAdditional struct {
	Service any
}

type Service struct {
	creator  func(Dic) any
	lifetime lifetime
	// initialized service for singleton
	// scope name for scoped
	additional any
}

func newSingleton(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		lifetime: singleton,
	}
}

func newScoped(scope ScopeID, creator func(Dic) any) Service {
	return Service{
		creator:    creator,
		lifetime:   scoped,
		additional: ScopedAdditional{Scope: scope},
	}
}

func newTransient(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		lifetime: transient,
	}
}

type Order int

const (
	DefaultOrder = iota
)

type ctorWrap struct {
	order Order
	wraps func(c Dic, s any) any
}

func newCtorWrap[T any](order Order, wrap func(c Dic, s T) T) ctorWrap {
	w := wrap
	return ctorWrap{order: order, wraps: func(c Dic, s any) any { return w(c, s.(T)) }}
}
