package ioc

type service struct {
	creator  func(Dic) any
	wraps    func(Dic, any)
	instance *any
}

func newService(creator func(Dic) any) service {
	var instance any
	return service{
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
