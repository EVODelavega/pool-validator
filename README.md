# pool-validator
Safe and easy request validation

## The goal

Sometimes you find yourself writing request validators for your handlers, and adding noise along the lines of:

```go
func (s *server) SomeHandler(ctx context.Context, req *api.Request, res *api.Response) error {
    if !s.validator(req, api.AllRequired) {
        return api.InvalidRequestErr
    }
}
```

The problem there is that these validators aren't all too flexible. I found myself writing more flexible validators, but then they required the use of a `sync.Pool` to be used safely. This is fine, but we're all human, and returning the validator back to the pool is something that is easily forgotten, and it makes the code look a bit messier than it needs to be (especially with more complex request objects):

```go
func (s *server) AnotherHandler(ctx context.Context, req *api.ComposedRequest, res *api.Response) error {
    userVal := s.userValPool.Get().(validator.User)
    fooVal := s.fooValPool.Get().(validator.Foo)
    defer func() {
        s.userValPool.Put(userVal)
        s.fooValPool.Put(fooVal)
    }()
    if !userVal.Validate(req.User) || !fooVal.Validate(req.Foo) {
        return api.InvalidRequestErr
    }
}
```

now this isn't bad looking, but as the name suggests: this request is a composed type. Before you know it, you'll find these `Get`, `Put` and `Validate` calls all over, not to mention the messy type-assertions and the `defer`'s.

When it comes to validating request data, those validators should be a fire-and-forget type of thing. That's where this package comes in.


### implementation

It's essentially a wrapper that holds a `sync.Pool` with your validators, and an `invoke` function where you perform your type assertions. You configure them once, assign them to a field on the `server` type (or wherever you need them), and the code above could be written like so:

```go
func (s *server) AnotherHandler(ctx context.Context, req *api.ComposedRequest, res *api.Response) error {
    if _, err := s.userVal.Validate(req.User); err != nil {
        log.Debug("%+v", err)
        return api.InvalidRequestErr
    }
    if _, err := s.fooVal.Validate(req.Foo); err != nil {
        log.Debug("%+v", err)
        return api.InvalidRequestErr
    }
}
```

or, depending on how elaborate your validators get:

```go
func (s *server) AnotherHandler(ctx context.Context, req *api.ComposedRequest, res *api.Response) error {
    if _, err := s.requestVal.Validate(req.User, req.Foo); err != nil {
        log.Debug("%+v", err)
        return api.InvalidRequestErr
    }
}
```

## Example

A simple example of how to use this validator can be found in the test. In essence, all you need to do is have a `New` function for the pool to use, and a callback for the type assertions. The first argument passed to this callback is the validator (passed as type `interface{}`), followed by varargs (again type `interface{}`).
This means you can write a single request validator that can validate all types in the `api` package (as per the examples above) like so:

```go
func invokeFunc(v interface{}, args ...interface{}) (interface{}, error) {
    tv := v.(*validator.RequestValidator)
    for _, a := range args {
        switch ta := a.(type) {
        case *api.Request:
            if !tv.ValidateRequest(ta) {
                return ta, api.InvalidRequestErr
            }
        case *api.User:
            if !tv.ValidateUser(ta) {
                return ta, api.InvalidRequestErr
            }
        // other cases here
        default:
            return nil, api.UnknownRequestErr
        }
    }
    // something like this
    return true, nil
}
```

## Where to go from here?

This is an initial commit, there's a lot more work that needs doing (including updating this README file, adding generic validators, and creating a roadmap)
