package ioc

import (
	"log"
	"reflect"
)

func typeKey[T any]() any {
	var service T
	return serviceKey(reflect.ValueOf(service))
}

func Get[T any](c Dic) T {
	var res T
	c.Inject(&res)
	return res
}

func GetServices[T any](c Dic) T {
	var res T
	c.InjectServices(&res)
	return res
}

func Scope(c Dic) Dic {
	return c.Scope()
}

func RegisterSingleton[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceResiterMutex.Lock()
	defer c.c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func RegisterScoped[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceResiterMutex.Lock()
	defer c.c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}

func RegisterTransient[T any](c Dic, creator func(c Dic) T) {
	c.c.serviceResiterMutex.Lock()
	defer c.c.serviceResiterMutex.Unlock()
	key := typeKey[T]()
	if _, ok := (*c.c.services)[key]; ok {
		var t T
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	(*c.c.services)[key] = service
}
