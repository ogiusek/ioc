package ioc

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sort"
	"sync"

	"github.com/ogiusek/lockset"
)

type ScopeID any

var (
	ScopeSingleton ScopeID = struct{}{}
	ScopeTransient ScopeID = struct{}{}
)

type serviceID any

type builder struct {
	wraps                 map[serviceID][]ctorWrap
	dependencies          map[reflect.Type][]reflect.Type
	services              map[serviceID]service
	eagerSingletonLoading bool
	eagerScoped           []serviceID
	scopes                map[ScopeID]struct{}
}

type Builder struct {
	b *builder
}

func NewBuilder() Builder {
	return Builder{
		b: &builder{
			wraps:                 map[serviceID][]ctorWrap{},
			dependencies:          map[reflect.Type][]reflect.Type{},
			services:              map[serviceID]service{},
			eagerSingletonLoading: true,
			scopes:                map[ScopeID]struct{}{},
		},
	}
}

func (b Builder) Clone() Builder {
	clonedWraps := make(map[serviceID][]ctorWrap)
	for k, v := range b.b.wraps {
		clonedWraps[k] = append([]ctorWrap{}, v...)
	}

	clonedDependencies := make(map[reflect.Type][]reflect.Type)
	for k, v := range b.b.dependencies {
		clonedDependencies[k] = append([]reflect.Type{}, v...)
	}

	clonedServices := make(map[serviceID]service)
	for k, v := range b.b.services {
		clonedServices[k] = v
	}

	clonedScopes := make(map[ScopeID]struct{})
	for k, v := range b.b.scopes {
		clonedScopes[k] = v
	}

	clonedEagerScoped := make([]serviceID, len(b.b.eagerScoped))
	copy(clonedEagerScoped, b.b.eagerScoped)

	clonedB := Builder{
		b: &builder{
			wraps:                 clonedWraps,
			dependencies:          clonedDependencies,
			services:              clonedServices,
			eagerSingletonLoading: b.b.eagerSingletonLoading,
			eagerScoped:           clonedEagerScoped,
			scopes:                clonedScopes,
		},
	}
	return clonedB
}

type ctorWraps []ctorWrap

func (a ctorWraps) Len() int      { return len(a) }
func (a ctorWraps) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ctorWraps) Less(i, j int) bool {
	if a[i].order != a[j].order {
		return a[i].order < a[j].order
	}
	return false
}

func (b Builder) Wrap(wrap func(Builder) Builder) Builder {
	return wrap(b)
}

func RegisterSingleton[Service any](b Builder, creator func(c Dic) Service) Builder {
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newSingleton(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	b.b.dependencies[reflect.TypeFor[Service]()] = nil
	return b
}

func RegisterScoped[Service any](b Builder, loading loading, scope ScopeID, creator func(c Dic) Service) Builder {
	if scope == ScopeSingleton {
		return RegisterSingleton(b, creator)
	}
	if scope == ScopeTransient {
		return RegisterTransient(b, creator)
	}
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newScoped(scope, func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	b.b.eagerScoped = append(b.b.eagerScoped, key)
	b.b.dependencies[reflect.TypeFor[Service]()] = nil
	return b
}

func RegisterTransient[Service any](b Builder, creator func(c Dic) Service) Builder {
	key := typeKey[Service]()
	if _, ok := b.b.services[key]; ok {
		var t Service
		log.Panicf("registered service already exists '%s'", reflect.TypeOf(t).String())
	}
	service := newTransient(func(c Dic) any { return creator(c) })
	b.b.services[key] = service
	b.b.dependencies[reflect.TypeFor[Service]()] = nil
	return b
}

// wraps with the smallest id are applied first
// wraps with the same order are applied randomly
func WrapService[Service any](b Builder, order Order, wrap func(c Dic, s Service) Service) Builder {
	key := typeKey[Service]()
	wraps := newCtorWrap(order, wrap)

	if _, ok := b.b.wraps[key]; !ok {
		b.b.wraps[key] = make([]ctorWrap, 0, 1)
	}

	b.b.wraps[key] = append(b.b.wraps[key], wraps)
	return b
}

func RegisterDependencies[Service any](b Builder, dependencies ...reflect.Type) Builder {
	tType := reflect.TypeFor[Service]()
	b.b.dependencies[tType] = append(b.b.dependencies[tType], dependencies...)
	return b
}

func RegisterDependency[Service any, Dependency any](b Builder) Builder {
	tType := reflect.TypeFor[Service]()
	b.b.dependencies[tType] = append(b.b.dependencies[tType], reflect.TypeFor[Dependency]())
	return b
}

// panics when attempting to regsiter ScopeSingleton or ScopeTransient
func (b Builder) RegisterScope(scope ScopeID) {
	if scope == ScopeSingleton {
		panic(ErrScopeDoesNotExist)
	}
	if scope == ScopeTransient {
		panic(ErrScopeDoesNotExist)
	}
	b.b.scopes[scope] = struct{}{}
}

func (b Builder) LazySingletonLoading() {
	b.b.eagerSingletonLoading = false
}

func validateDependencies(b Builder) error {
	visited := make(map[reflect.Type]bool)
	visiting := make(map[reflect.Type]bool)
	var dfsValidate func(
		currentType reflect.Type,
		path []reflect.Type,
	) error
	dfsValidate = func(
		currentType reflect.Type,
		path []reflect.Type,
	) error {
		visiting[currentType] = true
		path = append(path, currentType)

		currentDeps, ok := b.b.dependencies[currentType]
		if !ok {
			currentDeps = []reflect.Type{}
		}

		for _, depType := range currentDeps {
			// Check for circular dependency
			if visiting[depType] {
				// Cycle detected! Find the start of the cycle in the path.
				cycleStartIdx := -1
				for i, t := range path {
					if t == depType {
						cycleStartIdx = i
						break
					}
				}
				return errors.Join(
					ErrCircularDependency,
					fmt.Errorf("%v", append(path[cycleStartIdx:], depType)),
				)
				// return &ErrCircularDependency{Path: append(path[cycleStartIdx:], depType)}
			}

			// Check for missing dependency: If depType is not a key in the map, it's missing.
			// We only consider it missing if it's not already the currentType itself (self-dependency).
			// A self-dependency is allowed as long as it doesn't form a cycle with other nodes.
			if _, exists := b.b.dependencies[depType]; !exists && depType != currentType {
				return errors.Join(
					ErrMissingDependency,
					fmt.Errorf("missing \"%s\" required by \"%s\"", depType.String(), currentType.String()),
				)
			}

			// If the dependency has not been fully visited, recurse.
			if !visited[depType] {
				if err := dfsValidate(depType, path); err != nil {
					return err // Propagate error from recursion
				}
			}
		}

		// Done processing currentType and all its reachable dependencies.
		// Mark as visited and remove from visiting.
		delete(visiting, currentType)
		visited[currentType] = true

		return nil
	}

	// Iterate over each type as a potential starting point for DFS.
	for rootType := range b.b.dependencies {
		serviceKey := serviceKey(rootType)
		if _, ok := b.b.services[serviceKey]; !ok {
			return errors.Join(
				ErrServiceIsntRegistered,
				fmt.Errorf("\"%s\" isn't registered", rootType.String()),
			)
		}
		if visited[rootType] {
			continue // Already processed this branch
		}

		// Perform DFS from the current root.
		if err := dfsValidate(rootType, []reflect.Type{}); err != nil {
			return err // Propagate any error found
		}
	}

	return nil // No errors found
}

func (b Builder) Build() Dic {
	services := b.b.services
	if err := validateDependencies(b); err != nil {
		panic(err)
	}
	for key, service := range services {
		wraps, ok := b.b.wraps[key]
		if !ok || len(wraps) == 0 {
			continue
		}
		sort.Sort(ctorWraps(wraps))
		ctor := service.creator
		w := []ctorWrap(wraps)
		service.creator = func(d Dic) any {
			s := ctor(d)
			for _, wrap := range w {
				s = wrap.wraps(d, s)
			}
			return s
		}
		services[key] = service
	}
	eagerScopedServices := make(map[ScopeID][]serviceID, len(b.b.scopes))
	for scope := range b.b.scopes {
		eagerScopedServices[scope] = make([]serviceID, 0)
	}
	for serviceId, service := range services {
		if service.lifetime != scoped {
			continue
		}
		additional := service.additional.(scopedAdditional)
		if additional.Loading != EagerLoading {
			continue
		}
		scopeId := additional.Scope
		eagerScopedServices[scopeId] = append(eagerScopedServices[scopeId], serviceId)
	}
	c := Dic{
		c: &dic{
			serviceRegisterMutex: &sync.Mutex{},
			services:             services,
			eagerScopedServices:  eagerScopedServices,
			scopedCreateLockset:  lockset.New(),
			scopes:               map[ScopeID]map[serviceID]any{},
		},
	}
	for scopeId := range b.b.scopes {
		c.c.scopes[scopeId] = map[serviceID]any{}
	}
	if b.b.eagerSingletonLoading {
		for key, service := range b.b.services {
			if service.lifetime != singleton && (service.lifetime != scoped) {
				continue
			}
			if service.additional == nil {
				c.c.scopedCreateLockset.Lock(key)
				service.additional = singletonAdditional{
					Service: service.creator(c),
				}
				b.b.services[key] = service
				c.c.scopedCreateLockset.Unlock(key)
			}
		}
	}
	return c
}
