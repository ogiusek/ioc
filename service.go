package ioc

//

type loading int

const (
	EagerLoading loading = iota
	LazyLoading
)

//

type lifetime int

const (
	singleton lifetime = iota
	scoped
	transient
)

//

type scopedAdditional struct {
	Scope   ScopeID
	Loading loading
}

type singletonAdditional struct {
	Service any
	Loading loading
}

type service struct {
	creator  func(Dic) any
	lifetime lifetime
	// initialized service for singleton
	// scope name for scoped
	additional any
}

func newSingleton(creator func(Dic) any) service {
	return service{
		creator:  creator,
		lifetime: singleton,
	}
}

func newScoped(scope ScopeID, creator func(Dic) any) service {
	return service{
		creator:    creator,
		lifetime:   scoped,
		additional: scopedAdditional{Scope: scope},
	}
}

func newTransient(creator func(Dic) any) service {
	return service{
		creator:  creator,
		lifetime: transient,
	}
}

//

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
