# Slices and Arrays

In TypeScript, there's one type for ordered collections: `Array`. In Go, there are two: **arrays** and **slices**. You'll almost always use slices.

## Arrays: fixed size

A Go array has a fixed length that's part of its type:

```go
var a [3]int           // array of 3 ints, initialized to [0, 0, 0]
b := [3]string{"a", "b", "c"}
```

The key difference from TypeScript: `[3]int` and `[5]int` are **different types**. You can't pass a `[3]int` to a function that expects `[5]int`. This makes arrays rigid and rarely used directly.

In TypeScript, the closest equivalent would be a fixed-length tuple:

```typescript
const a: [number, number, number] = [1, 2, 3];
```

But even TypeScript tuples are more flexible than Go arrays.

## Slices: what you'll actually use

A slice is a dynamically-sized view over an array. This is what you'll reach for 99% of the time — it's the Go equivalent of a JavaScript `Array`.

```go
// Declare a slice (no size in the brackets)
todos := []Todo{
	{ID: 1, Title: "Learn Go", Completed: false},
	{ID: 2, Title: "Build a web server", Completed: false},
}
```

Notice `[]Todo` vs `[2]Todo` — no number in the brackets means it's a slice.

## Common operations

Here's how familiar array operations map from TypeScript to Go:

### Append

```typescript
// TypeScript
todos.push({ id: 3, title: "New todo" });
```

```go
// Go
todos = append(todos, Todo{ID: 3, Title: "New todo"})
```

`append` doesn't modify the original slice — it returns a new one. You must reassign it. This catches people coming from TypeScript where `.push()` mutates in place.

### Length

```typescript
// TypeScript
todos.length
```

```go
// Go
len(todos)
```

### Access by index

Same in both languages:

```go
first := todos[0]
```

### Iterate

```typescript
// TypeScript
todos.forEach((todo, index) => { ... });
```

```go
// Go
for index, todo := range todos {
	// ...
}
```

`range` returns two values: the index and a copy of the element. If you don't need the index, use `_` to discard it:

```go
for _, todo := range todos {
	fmt.Println(todo.Title)
}
```

### Slice (sub-array)

```typescript
// TypeScript
todos.slice(1, 3)
```

```go
// Go
todos[1:3]
```

### Delete

There's no `.splice()` equivalent built into the language, but Go's standard library has a [`slices`](https://pkg.go.dev/slices) package with a `Delete` function:

```typescript
// TypeScript
todos.splice(index, 1);
```

```go
// Go
import "slices"
todos = slices.Delete(todos, index, index+1) // (slice, start index, end index)
```

Like `append`, it returns a new slice — you must reassign it.

### Filter

This is where Go feels verbose compared to TypeScript. There's no built-in `.filter()`:

```typescript
// TypeScript
const completed = todos.filter(t => t.completed);
```

```go
// Go
var completed []Todo
for _, t := range todos {
	if t.Completed {
		completed = append(completed, t)
	}
}
```

No one-liners here. You write the loop.

## One gotcha: range copies

When you iterate with `range`, each element is a **copy**, not a reference. This is different from TypeScript's `forEach`:

```go
for _, todo := range todos {
	todo.Completed = true // modifies the copy, not the original!
}
```

To modify elements in place, use the index:

```go
for i := range todos {
	todos[i].Completed = true // modifies the original
}
```

This trips up every TypeScript dev at least once.

## Where you'll see this in the project

In our todo API, the in-memory store is a slice:

```go
var todos = []Todo{
	{ID: 1, Title: "Learn Go", Completed: false},
	{ID: 2, Title: "Build a web server", Completed: false},
	{ID: 3, Title: "Write unit tests", Completed: false},
}
```

We use `append` to add new todos, `range` to search by ID, and slice re-assignment to delete items. You'll see all of these in the upcoming route lessons.
