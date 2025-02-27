package ioc

type Service struct {
	creator  func(Dic) any
	lifetime lifetime
}

func newSingleton(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		lifetime: singleton,
	}
}

func newScoped(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		lifetime: scoped,
	}
}

func newTransient(creator func(Dic) any) Service {
	return Service{
		creator:  creator,
		lifetime: transient,
	}
}
