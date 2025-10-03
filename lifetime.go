package ioc

type lifetime int

const (
	singleton lifetime = iota
	scoped
	transient
)
