# Interfaces

Interfaces in Go look similar to TypeScript interfaces on the surface, but they work very differently. Understanding the difference is one of the more important mental shifts when coming from TypeScript.

## TypeScript: explicit, opt-in

In TypeScript, an interface is a contract you **explicitly declare**. A class must announce that it satisfies it:

```typescript
interface TodoRepository {
    getAll(): Promise<Todo[]>
    getByID(id: number): Promise<Todo | null>
}

class DatabaseRepo implements TodoRepository {  // ← explicit declaration
    async getAll() { ... }
    async getByID(id: number) { ... }
}

class FakeRepo implements TodoRepository {  // ← also explicit
    async getAll() { ... }
    async getByID(id: number) { ... }
}
```

`implements` is a promise you make. TypeScript verifies it. But critically: the class knows about the interface.

## Go: implicit, structural

In Go, there is no `implements` keyword. A type satisfies an interface automatically if it has the right methods — the type never needs to mention the interface at all.

```go
type TodoStorer interface {
    GetAll(ctx context.Context) ([]Todo, error)
    GetByID(ctx context.Context, id int64) (*Todo, error)
}

// Real implementation — never mentions TodoStorer
type TodoStore struct{ db *bun.DB }
func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) { ... }
func (s *TodoStore) GetByID(ctx context.Context, id int64) (*Todo, error) { ... }

// Test fake — also never mentions TodoStorer
type inMemoryStore struct{ todos map[int64]*Todo }
func (s *inMemoryStore) GetAll(ctx context.Context) ([]Todo, error) { ... }
func (s *inMemoryStore) GetByID(ctx context.Context, id int64) (*Todo, error) { ... }
```

Both types satisfy `TodoStorer` purely because their method signatures match. The compiler checks this at the point of **use**, not definition.

The point of use looks like this:

```go
// This function accepts any type that satisfies TodoStorer
func NewHandler(s store.TodoStorer) *Handler {
    return &Handler{store: s}
}

// Both of these compile:
NewHandler(store.NewTodoStore(db))   // real DB
NewHandler(newInMemoryStore())       // test fake
```

The compiler verifies that each argument has all the required methods at the call site. If `inMemoryStore` is missing a method or has the wrong signature, you get a compile error — not a runtime panic.

## Why the difference matters

In TypeScript, the interface and the implementation are coupled by `implements`. The class file must import the interface.

In Go, they are fully decoupled. `inMemoryStore` in `handlers_test.go` satisfies `store.TodoStorer` without importing the `store` package at all — as long as the methods match. You can satisfy interfaces from packages you didn't write, can't modify, or don't even know about.

This is how Go's standard library works. `io.Reader` is just:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

`os.File`, `bytes.Buffer`, `net.Conn`, `http.Response.Body` — none of them say `implements io.Reader`. They just happen to have a `Read` method with that signature. They are all `io.Reader`s. Any function that accepts `io.Reader` works with all of them.

## Interface values

When you assign a concrete type to an interface variable, Go stores two things internally: the concrete type and a pointer to the value. This is called an **interface value**.

```go
var s store.TodoStorer       // s is nil — no type, no value
s = newInMemoryStore()       // s now holds (*inMemoryStore, pointer-to-store)
```

This is why the nil interface check exists in Go:

```go
if s == nil { ... }   // true only if both type and value are nil
```

A common gotcha: a nil pointer of a concrete type assigned to an interface is NOT a nil interface:

```go
var fake *inMemoryStore = nil  // typed nil
var s store.TodoStorer = fake  // s has type *inMemoryStore, value nil
s == nil                       // false — has a type, even though value is nil
```

You will not hit this in normal code, but it trips up everyone eventually.

## Interfaces are usually small

Go interfaces tend to be small — often a single method. The standard library is full of them:

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Stringer interface { String() string }
type error interface { Error() string }
```

`error` itself is an interface. That is why any type with an `Error() string` method is a valid Go error — no explicit declaration needed.

The Go proverb: **"the bigger the interface, the weaker the abstraction."** Small interfaces are more reusable because more types can satisfy them. Our `TodoStorer` with 5 methods is on the larger side — acceptable for a repository, but not something you'd compose freely.

## When to define an interface

A Go-specific rule of thumb: **define interfaces at the point of use, not the point of implementation**.

In TypeScript, you often define the interface in the same file as (or near) the class that implements it. In Go, the convention is to define the interface where it's consumed. In our project:

- `store/todo.go` defines `TodoStore` (the implementation)
- `store/todo.go` also defines `TodoStorer` (the interface) — but only because the store and its interface are tightly related here

If the handlers file were in a different package, the interface would typically live there, not in `store/`. The consumer defines what it needs; the producer just happens to have it. This keeps packages loosely coupled.

## Embedding interfaces

Go interfaces can embed other interfaces, composing them:

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }

type ReadWriter interface {
    Reader  // embeds Reader
    Writer  // embeds Writer
}
```

`ReadWriter` requires both `Read` and `Write`. This is composition rather than inheritance — the same idea you know from TypeScript's `interface C extends A, B`.

## The empty interface

Before Go 1.18, you'd see `interface{}` everywhere:

```go
func Print(v interface{}) { ... }  // accepts anything
```

Since Go 1.18, `any` is an alias for `interface{}` and is preferred:

```go
func Print(v any) { ... }  // same thing, cleaner syntax
```

`any` satisfies all interfaces because it has no method requirements. It's Go's equivalent of TypeScript's `unknown` (or the less safe `any`). Avoid it where possible — using `any` loses type safety.
