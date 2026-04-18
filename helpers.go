package ioc

// Transient is just a factory which can be registered
type Transient[Service any] func() Service

// For every service its lazy getter is automatically registered
type Lazy[Service any] func() Service

// pkg is an interface recommended to use
type Pkg func(b Builder)

//

func NewPkg(r func(b Builder)) Pkg { return r }

func NewPkgT[Config any](r func(Builder, Config)) func(Config) Pkg {
	return func(c Config) Pkg { return func(b Builder) { r(b, c) } }
}
