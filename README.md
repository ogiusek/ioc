# ioc

this package defines "dependency injection container"

it is type safe

## example usage

```go
package main

import "github.com/ogiusek/ioc"

type Session struct {
    Name string
}

func main(){
	c := ioc.NewContainer()
	ioc.RegisterSingleton(c, func(d ioc.Dic) int {
        return 1
    })
	ioc.RegisterScoped(c, func(d ioc.Dic) *Session {
		return &Session{}
	})

	scope := ioc.Scope(c)
    session := ioc.Get[*Session](scope)
    *session = Session{
        Name: "john",
    }

    session = ioc.Get[*Session](scope)
    // session isn't nil
    if session != nil {
        log.Printf("user name is '%s'", session.Name)
        // this will print:
        // "user name is '%s'"
    }

}
```