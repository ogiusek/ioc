package ioc

import "errors"

var (
	ErrScopeDoesNotExist         error = errors.New("scope does not exists")
	ErrScopeIsNotInitialized     error = errors.New("scope is not initialized")
	ErrScopeIsAlreadyInitialized error = errors.New("scope is already initialized")

	ErrIsntPointer           error = errors.New("isn't a pointer")
	ErrIsntPointerToStruct   error = errors.New("isn't a pointer to a struct")
	ErrServiceIsntRegistered error = errors.New("service isn't registered")

	ErrHasToRegisterServiceBeforeDependencies error = errors.New("has to register service before registering its dependencies")

	ErrCircularDependency error = errors.New("circular dependency")
	ErrMissingDependency  error = errors.New("missing dependency")
)
