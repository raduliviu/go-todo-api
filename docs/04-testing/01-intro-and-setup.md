# Testing in Go: Intro and Setup

So far we've been testing the API manually with `curl`. That works, but it doesn't scale — as the codebase grows, you can't curl your way through every scenario after every change. Automated tests do that for you.

This chapter covers **HTTP handler testing**: sending fake HTTP requests to your handlers and asserting the responses, all without starting a real server.

## How Go testing works

Go has testing built in. No Jest, no Mocha, no third-party test runner needed.

- Test files must end in `_test.go`. The compiler ignores them in normal builds.
- Test functions must start with `Test` and accept `*testing.T` as their only argument.
- You run them with `go test ./...`

```go
func TestSomething(t *testing.T) {
    // test code here
}
```

This is roughly equivalent to:

```typescript
// TypeScript / Jest
test('something', () => {
    // test code here
})
```

The `t` parameter is your handle for marking failures and logging. The two methods you'll use most:

- `t.Errorf(...)` — marks the test as failed but keeps running (like `expect().toBe()` in Jest)
- `t.Fatalf(...)` — marks the test as failed and stops immediately (useful when a failed step makes everything after it meaningless)

## Package choice: white-box vs black-box

When you create `handlers_test.go`, the first line is the package declaration. You have two options:

```go
package main       // white-box: same package, can access unexported identifiers
package main_test  // black-box: separate package, only sees exported identifiers
```

We use `package main` here because `todos` and `setupRouter` are unexported (lowercase). A `package main_test` file would be a completely separate package and couldn't see them at all.

This is a Go-specific concept. In TypeScript, everything you `export` is visible to importers — there's no concept of "same package" access. In Go, anything lowercase is private to its package, and test files can opt in to that same package.

## TestMain: package-level setup

Go has a special function called `TestMain` that, if present, runs instead of running tests directly. You do your setup, call `m.Run()` to actually run the tests, and optionally clean up after:

```go
func TestMain(m *testing.M) {
    gin.SetMode(gin.TestMode)
    m.Run()
}
```

`gin.SetMode(gin.TestMode)` tells Gin not to print its `[GIN-debug]` route registration output during tests. Without it, every `go test` run would be noisy.

This is similar to Jest's `beforeAll` at the global level — it runs once for the entire test package.

## What's next

With the test file set up, the next lesson introduces the two helper functions that make every test possible: `resetTodos` and `performRequest`.
