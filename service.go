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

func (s *Service) wrap(wrap ctorWraps) {
	ctor := s.creator
	s.creator = func(c Dic) any { return wrap.wraps(c, ctor(c)) }
}

type ctorWraps struct {
	wraps func(c Dic, s any) any
}

func newCtorWrap[T any](wrap func(c Dic, s T) T) ctorWraps {
	return ctorWraps{wraps: func(c Dic, s any) any { return wrap(c, s.(T)) }}
}

func (wraps *ctorWraps) wrap(wrap ctorWraps) {
	w := wraps.wraps
	wraps.wraps = func(c Dic, s any) any { return w(c, wrap.wraps(c, s)) }
}
