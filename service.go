package ioc

type Service struct {
	creator  func(Dic) any
	wraps    func(Dic, any)
	instance *any
}

func newSingleton(creator func(Dic) any) Service {
	var instance any
	return Service{
		creator:  creator,
		wraps:    func(d Dic, a any) {},
		instance: &instance,
	}
}

type ctorWrap struct {
	wraps func(c Dic, s any)
}

func newCtorWrap[T any](wrap func(c Dic, s T)) ctorWrap {
	w := wrap
	return ctorWrap{wraps: func(c Dic, s any) { w(c, s.(T)) }}
}
