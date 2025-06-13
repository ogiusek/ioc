# **ioc \- a simple and ergonomic dependency injection container for go**

ioc is a lightweight and opinionated dependency injection (di) container for go, designed to minimize boilerplate and maximize developer experience. it provides a straightforward way to manage and inject dependencies into your go applications, promoting loosely coupled and testable code.

## **what is dependency injection?**

dependency injection is a design pattern that allows for the removal of hard-coded dependencies among components. instead of a component creating its own dependencies, they are provided (injected) from an external source, typically a di container. this leads to:

* **reduced coupling:** components become less reliant on specific implementations, making your codebase more flexible and easier to maintain.
* **improved testability:** dependencies can be easily mocked or stubbed during testing, allowing for isolated unit tests.
* **increased reusability:** components become more generic and can be reused in different contexts.
* **better code organization:** centralized dependency management clarifies how components are wired together.

## **why ioc?**

while other go di packages exist, ioc prioritizes:

* **no boilerplate:** define your services and let ioc handle the wiring. you won't find yourself writing extensive setup code or complex configuration files.
* **simplicity:** the api is intuitive and easy to understand, even for developers new to dependency injection.
* **developer experience:** ioc aims to get out of your way and let you focus on writing business logic. features like type-safe get\[t\] and automatic injectservices via struct tags make common tasks a breeze.
* **go-idiomatic:** leverages go's features like generics and struct tags to provide a natural and efficient experience.

## **how it differs from other packages**

many di containers for go often require:

* **code generation:** generating code to resolve dependencies, which can add complexity to the build process and sometimes obscure the underlying logic.
* **complex configuration:** using yaml, toml, or complex go struct configurations to define dependencies, which can be verbose and less type-safe.
* **reflection-heavy apis:** relying heavily on reflection for injection, which can sometimes be less performant and lead to runtime errors if not handled carefully.

ioc avoids these by:

* **runtime resolution:** dependencies are resolved at runtime, eliminating the need for code generation.
* **function-based registration:** services are registered using simple go functions (creators), providing full type safety and ide support.
* **targeted reflection:** reflection is used judiciously (e.g., for injectservices via struct tags) to enhance developer convenience without sacrificing clarity or performance.

## **installation**

go get github.com/ogiusek/ioc

## **usage**

### **basic registration and injection**

let's start with a simple example:
```go
package main

import (
	"fmt"
	"github.com/ogiusek/ioc"
)

// define our services
type logger interface {
	log(msg string)
}

type consolelogger struct{}

func (c \*consolelogger) log(msg string) {
	fmt.printf("\[consolelogger\] %s\\n", msg)
}

type greeter struct {
	logger logger \`inject:"1"\` // ioc will inject the logger here
}

func (g \*greeter) greet(name string) {
	g.logger.log(fmt.sprintf("hello, %s\!", name))
}

func main() {
	// 1\. create a new di container
	dic := ioc.newcontainer()

	// 2\. register services
	ioc.registersingleton(dic, func(c ioc.dic) logger {
		return \&consolelogger{}
	})

	// register greeter as a transient service (new instance every time)
	ioc.registertransient(dic, func(c ioc.dic) \*greeter {
		return \&greeter{}
	})

	// 3\. get and use services
	greeter := ioc.get\[\*greeter\](dic)
	greeter.greet("world")

	// another way to get services: injectservices
	type myapp struct {
		greeter \*greeter \`inject:"1"\`
	}

	app := ioc.getservices\[myapp\](dic)
	app.greeter.greet("ioc user")
}
```

### **service lifecycles**

ioc supports three main service lifecycles:

* **singleton:** a single instance of the service is created and shared across the entire container and its scopes.
  ioc.registersingleton(dic, func(c ioc.dic) myservice {
      return \&myserviceimpl{}
  })

* **scoped:** a new instance of the service is created for each new scope. within a single scope, the same instance is returned.
  ioc.registerscoped(dic, func(c ioc.dic) myscopedservice {
      return \&myscopedserviceimpl{}
  })

* **transient:** a new instance of the service is created every time it's requested.
  ioc.registertransient(dic, func(c ioc.dic) mytransientservice {
      return \&mytransientserviceimpl{}
  })

### **scopes**

scopes allow you to manage the lifetime of services within a specific context (e.g., an http request, a user session). services registered as scoped will have a new instance created for each new scope.
```go
package main

import (
	"fmt"
	"github.com/ogiusek/ioc"
)

type requestscopedservice struct {
	id int
}

var currentrequestid int

func newrequestscopedservice() \*requestscopedservice {
	currentrequestid++
	return \&requestscopedservice{id: currentrequestid}
}

func main() {
	rootdic := ioc.newcontainer()
	ioc.registerscoped(rootdic, func(c ioc.dic) \*requestscopedservice {
		return newrequestscopedservice()
	})

	// first request scope
	request1dic := ioc.scope(rootdic)
	svc1\_req1 := ioc.get\[\*requestscopedservice\](request1dic)
	svc2\_req1 := ioc.get\[\*requestscopedservice\](request1dic)
	fmt.printf("request 1, service 1 id: %d\\n", svc1\_req1.id) // same id
	fmt.printf("request 1, service 2 id: %d\\n", svc2\_req1.id) // same id

	// second request scope
	request2dic := ioc.scope(rootdic)
	svc1\_req2 := ioc.get\[\*requestscopedservice\](request2dic)
	svc2\_req2 := ioc.get\[\*requestscopedservice\](request2dic)
	fmt.printf("request 2, service 1 id: %d\\n", svc1\_req2.id) // new id
	fmt.printf("request 2, service 2 id: %d\\n", svc2\_req2.id) // same id as svc1\_req2
}
```

### **injecting dependencies**

ioc provides two primary ways to inject dependencies:

1. **dic.inject(servicepointer any)**: injects a single service into a pointer variable.
```go
   var mylogger logger
   err := dic.inject(\&mylogger)
   if err \!= nil {
       // handle error, e.g., ioc.errserviceisntregistered
   }
   mylogger.log("injected directly\!")
```

2. **dic.injectservices(services any)**: injects multiple dependencies into a struct by inspecting fields with the inject:"1" tag.
```go
   type myconsumers struct {
       logger logger \`inject:"1"\`
       db     db     \`inject:"1"\`
   }

   var consumers myconsumers
   err := dic.injectservices(\&consumers)
   if err \!= nil {
       // handle error, e.g., ioc.errisntpointertostruct
   }
   consumers.logger.log("injected via struct\!")
```

### **retrieving services (panicking vs. error handling)**

ioc offers two styles for retrieving services:

* **get\[t\](c dic) t**: this generic function is designed for convenience and will **panic** if the service of type t is not registered or if injection fails. this is ideal for applications where you expect services to always be available after initial setup.
```go
  logger := ioc.get\[logger\](dic) // panics if logger is not registered
```

* **inject(servicepointer any)**: as shown above, this method returns an error if the service cannot be injected. this is suitable when you need to handle potential missing dependencies explicitly.

similarly, for injecting multiple services into a struct:

* **getservices\[t\](c dic) t**: creates a new instance of t, injects its dependencies, and **panics** on failure.
```go
  myapp := ioc.getservices\[myapp\](dic) // panics if dependencies cannot be injected
```

* **injectservices(services any)**: injects into an existing struct and returns an error on failure.

### **error handling**

ioc defines several errors that can be returned by its methods:

* errisntpointer: returned by inject when the provided argument is not a pointer.
* errisntpointertostruct: returned by injectservices when the provided argument is not a pointer to a struct.
* errserviceisntregistered: returned by inject when the requested service type is not registered in the container.

additionally, registersingleton, registerscoped, and registertransient will **panic** if you attempt to register a service that has already been registered with the same type. this helps catch misconfigurations early.

## **api reference**

### **type dic**

the dic type represents the dependency injection container.
```go
type dic struct {
	c \*dic // internal, not for direct use
}
```

#### **newcontainer() dic**

creates and returns a new empty dependency injection container.

#### **(c dic) scope() dic**

creates and returns a new child scope from the current container. services registered as scoped will have a new instance created within this new scope.

#### **(c dic) inject(servicepointer any) error**

* servicepointer: a pointer to the variable where the service instance should be injected.
* **returns:**
  * nil if the service is successfully injected.
  * errisntpointer if servicepointer is not a pointer.
  * errserviceisntregistered if a service of the required type is not registered in the container.

#### **(c dic) injectservices(services any) error**

* services: a pointer to a struct. fields within this struct tagged with inject:"1" will be automatically populated with instances from the container.
* **returns:**
  * nil if all tagged services are successfully injected.
  * errisntpointertostruct if services is not a pointer to a struct.
  * any error returned by c.inject() if a specific service injection fails.

### **generic functions**

these functions provide a convenient, type-safe way to interact with the container. they will panic on failure.

#### **get\[t any\](c dic) t**

* c: the dic container.
* **returns:** an instance of type t.
* **panics:** if t is not registered or if c.inject() returns an error.

#### **getservices\[t any\](c dic) t**

* c: the dic container.
* **returns:** a new instance of type t with its tagged dependencies injected.
* **panics:** if t is not a struct type, or if c.injectservices() returns an error.

#### **scope(c dic) dic**

* c: the dic container.
* **returns:** a new child scope. this is a convenience wrapper around c.scope().

#### **registersingleton\[t any\](c dic, creator func(c dic) t)**

* c: the dic container.
* creator: a function that creates an instance of type t. this function receives the dic itself, allowing for nested dependency resolution.
* **panics:** if a service of type t has already been registered.

#### **registerscoped\[t any\](c dic, creator func(c dic) t)**

* c: the dic container.
* creator: a function that creates an instance of type t.
* **panics:** if a service of type t has already been registered.

#### **registertransient\[t any\](c dic, creator func(c dic) t)**

* c: the dic container.
* creator: a function that creates an instance of type t.
* **panics:** if a service of type t has already been registered.

## **contributing**

This isn't something i expect to happen therefor i do not has policy for this yet

## **license**

This project is licensed under the MIT License. See the [LICENSE](https://opensource.org/licenses/MIT) file for details.
