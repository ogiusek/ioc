package ioc

type ScopedAdditional struct {
	Scope ScopeID
}

type SingletonAdditional struct {
	Service any
}

type Service struct {
	creator  func(Dic) any
	wraps    func(Dic, any)
	lifetime lifetime
	// initialized service for singleton
	// scope name for scoped
	additional any
}

func newSingleton(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		wraps:    func(d Dic, a any) {},
		lifetime: singleton,
	}
}

func newScoped(scope ScopeID, creator func(Dic) any) Service {
	return Service{
		creator:    creator,
		wraps:      func(d Dic, a any) {},
		lifetime:   scoped,
		additional: ScopedAdditional{Scope: scope},
	}
}

func newTransient(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		wraps:    func(d Dic, a any) {},
		lifetime: transient,
	}
}

type Order int

const (
	DefaultOrder Order = iota
)

type ctorWrap struct {
	order Order
	wraps func(c Dic, s any)
}

func newCtorWrap[T any](order Order, wrap func(c Dic, s T)) ctorWrap {
	w := wrap
	return ctorWrap{order: order, wraps: func(c Dic, s any) { w(c, s.(T)) }}
}
