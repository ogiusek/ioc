package ioc

type Pkg interface {
	Register(b Builder) Builder
}
