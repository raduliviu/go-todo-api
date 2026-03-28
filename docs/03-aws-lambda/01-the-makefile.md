# The Makefile

SAM needs to compile your Go code into a Linux binary before it can deploy it to Lambda. The Makefile tells SAM exactly how to do that.

## Why a Makefile?

SAM has built-in support for several runtimes (Python, Node, Java). Go's managed runtime (`go1.x`) was deprecated in 2023. We now use `provided.al2023` — a bare Amazon Linux 2023 environment where we supply our own binary. Since SAM has no built-in knowledge of how to build this, we teach it with a Makefile target.

## The target

```makefile
build-TodoFunction:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap .
```

### `build-TodoFunction`

The target name SAM looks for when building `TodoFunction` from `template.yaml`. The convention is `build-` followed by the function's logical name.

### `GOOS=linux GOARCH=amd64`

Environment variables passed to the Go compiler. Your machine runs macOS (Darwin), but Lambda runs on Linux x86_64. Go can cross-compile — these two variables tell it to produce a Linux binary regardless of the host OS.

### `CGO_ENABLED=0`

Disables CGO, which allows Go code to call C libraries. C code is platform-specific and can't be cross-compiled cleanly. Setting this to `0` forces a pure Go binary — required because Lambda's Linux environment doesn't have the C runtime libraries your Mac has.

### `-o $(ARTIFACTS_DIR)/bootstrap`

Names the output binary `bootstrap` and places it in a directory SAM injects at build time. Lambda's `provided.al2023` runtime specifically looks for a file named `bootstrap` to execute — if it doesn't exist or isn't executable, Lambda returns a `Runtime.InvalidEntrypoint` error. See [Custom runtime entry point](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html) in the AWS docs.

### `.`

Build the Go package in the current directory.
