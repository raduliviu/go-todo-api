# Pointers

In TypeScript, you never think about pointers. Objects and arrays are always passed by reference, and primitives are passed by value. JavaScript handles this behind the scenes.

In Go, **everything is passed by value** — structs, ints, strings, all of it. When you pass a struct to a function, Go copies the entire thing. This is where pointers come in: they let you pass a *reference* to a value instead of copying it.

## The basics

A pointer holds the memory address of a value. Two operators to know:

- `&` — "address of" — gives you a pointer to a value
- `*` — "dereference" — gives you the value a pointer points to

```go
name := "Gopher"
ptr := &name     // ptr is a *string (pointer to a string)
fmt.Println(ptr) // 0xc0000140a0 (a memory address)
fmt.Println(*ptr) // "Gopher" (the value at that address)
```

## Why does this matter?

Consider this function:

```go
type Todo struct {
	Title     string
	Completed bool
}

func markComplete(t Todo) {
	t.Completed = true
}

func main() {
	todo := Todo{Title: "Learn Go", Completed: false}
	markComplete(todo)
	fmt.Println(todo.Completed) // false — the original didn't change!
}
```

`markComplete` received a *copy* of `todo`. It modified the copy, and the original was untouched. In TypeScript, this would have worked because objects are passed by reference.

To modify the original in Go, use a pointer:

```go
func markComplete(t *Todo) {
	t.Completed = true
}

func main() {
	todo := Todo{Title: "Learn Go", Completed: false}
	markComplete(&todo)
	fmt.Println(todo.Completed) // true
}
```

`*Todo` means "pointer to a Todo." `&todo` means "give me the address of todo." Now the function operates on the original value, not a copy.

## Reading pointer syntax

You'll see pointers everywhere in Go code. Here's how to read them:

| Syntax | Meaning |
|--------|---------|
| `*Todo` | The type "pointer to a Todo" |
| `&todo` | "The address of `todo`" |
| `*ptr` | "The value that `ptr` points to" |

Yes, `*` has two meanings depending on context — in a type declaration it means "pointer to", and as an operator it means "dereference." This is confusing at first, but you get used to it.

## When you'll see pointers in this project

### Gin's Context

Every Gin handler takes `*gin.Context`:

```go
func getTodos(c *gin.Context) {
	c.JSON(http.StatusOK, todos)
}
```

Why a pointer? The `Context` struct is large — it holds the request, response writer, middleware state, and more. Copying it on every function call would be wasteful. A pointer lets every handler work with the same Context object without copying it.

### JSON decoding

When you parse a JSON request body, you pass a pointer so the decoder can fill in your struct:

```go
var newTodo Todo
c.ShouldBindJSON(&newTodo) // pass address so Gin can write into newTodo
```

Without `&`, Gin would receive a copy and your `newTodo` variable would stay empty.

## The TypeScript mental model

Think of it this way:

- **Go by default** = everything behaves like a TypeScript primitive (copied on assignment/pass)
- **Go with pointers** = behaves like a TypeScript object (shared reference)

The difference is that in Go, you choose explicitly. In TypeScript, the language chooses for you based on the type.
