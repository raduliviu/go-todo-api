# Handler Tests: What Each One Tests and Why

This lesson walks through the five handler tests and highlights the interesting decisions in each.

## TestGetTodos

The simplest test — no mutations, one case.

```go
func TestGetTodos(t *testing.T) {
    resetTodos()
    router := setupRouter()

    w := performRequest(router, http.MethodGet, "/todos", nil)

    if w.Code != http.StatusOK {
        t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
    }

    var got []Todo
    if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
        t.Fatalf("could not unmarshal response: %v", err)
    }

    if len(got) != 3 {
        t.Errorf("count: got %d, want 3", len(got))
    }
}
```

We check the count rather than doing a deep comparison of the full slice. Either approach is valid. Checking count is lighter and makes the assertion intent clear: "this endpoint returns all todos".

Notice `nil` as the last argument to `performRequest` — GET requests have no body.

## TestGetTodoByID

Three cases: found, not found, invalid ID. This is the first use of the table-driven pattern with optional fields.

The interesting case is checking that the returned todo matches exactly:

```go
if got != *tc.wantTodo {
    t.Errorf("body: got %+v, want %+v", got, *tc.wantTodo)
}
```

`tc.wantTodo` is `*Todo` (a pointer). We dereference it with `*tc.wantTodo` to compare the values, not the pointers. Two pointers are only equal if they point to the same address — that's not what we want here.

For error cases, `tc.wantTodo` is `nil`, so the whole block is skipped.

## TestCreateTodo

The first mutating test. `POST /todos` appends to the `todos` slice, so `resetTodos()` lives inside the subtest loop.

The ID assertion is subtle:

```go
wantTodo: &Todo{ID: 4, Title: "Deploy to prod", Completed: false},
```

The handler assigns `newTodo.ID = len(todos) + 1`. After `resetTodos()`, there are 3 todos, so the next ID is 4. This only works reliably because we reset before every case — without that reset, a previous case creating a todo would change `len(todos)` and break the ID expectation.

For the malformed JSON case, we just check the status code:

```go
{
    name:       "malformed JSON returns 400",
    body:       []byte(`{bad json}`),
    wantStatus: http.StatusBadRequest,
},
```

`wantTodo` is `nil` so body checking is skipped. We don't assert the exact error message here because it comes from Go's JSON parser — it's an internal error string we don't control and that could change.

## TestUpdateTodoByID

The most cases of any handler. The most interesting one:

```go
{
    name:       "client cannot override id",
    path:       "/todos/1",
    body:       []byte(`{"id": 999, "title": "sneaky", "completed": false}`),
    wantStatus: http.StatusOK,
    wantTodo:   &Todo{ID: 1, Title: "sneaky", Completed: false},
},
```

The body sends `"id": 999`, but the handler always overwrites it: `updatedTodo.ID = id` (where `id` comes from the URL path). The response ID must be 1, not 999. This test documents and enforces that invariant. Without it, a future change to the handler could accidentally allow clients to rewrite IDs and no test would catch it.

This is the real purpose of tests: not just verifying that things work today, but catching regressions when someone changes something later.

## TestDeleteTodoByID

The 204 No Content case is different from all others:

```go
if tc.wantStatus == http.StatusNoContent && w.Body.Len() != 0 {
    t.Errorf("expected empty body for 204, got: %s", w.Body.String())
}
```

By HTTP spec, 204 responses must have no body. The handler calls `c.Status(http.StatusNoContent)` — no `c.JSON(...)`. We verify the body is actually empty rather than assuming. This is especially worth checking because it's easy to accidentally add a body to a delete response.

Never call `json.Unmarshal` on a 204 response body — there's nothing to parse and you'll get an error.

## Running the full suite

```bash
go test ./... -v
```

All 13 tests should pass. The `-v` flag shows every subtest by name. When one fails in the future, the output will tell you exactly which case broke and what values differed.

To run just one test function:

```bash
go test ./... -run TestGetTodoByID
```

To bypass Go's test result cache (Go caches passing results and skips re-running them if nothing changed):

```bash
go test ./... -count=1
```
