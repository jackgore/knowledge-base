# session

* `manager.go` 
    * defines the `Manager` interface for creating and retrieving user sessions.
    * this is a public interface that should remain relatively stable.
* `session.go`
   * defines the `Session` structure that is used by the session manager interface.
* `managers/`
    * This folder defines individual implementations of the the `Manager` interface.

## implementations

## smmanager
* This implementation is based on the `sync` packages Map type.
* Thread safe map, that is most performant for a ready heavy workload.

Used like this:
```
sm, err := NewSMManager("knowledge_base", 3600*24*365)
if err != nil {
    return nil, err
}
```
