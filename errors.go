package ioc

import "errors"

var (
	ErrIsntPointer         error = errors.New("isn't a pointer")
	ErrIsntPointerToStruct error = errors.New("isn't a pointer to a struct")

	ErrServiceIsntRegistered error = errors.New("service isn't registered")
	ErrCircularDependency    error = errors.New("circular dependency")
)
