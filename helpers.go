package ioc

type Lazy[Service any] func() Service

type Pkg interface {
	Register(b Builder)
}
