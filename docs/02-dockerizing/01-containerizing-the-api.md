# Containerizing the API

Our API runs locally with `go run .`, but that requires Go to be installed. Docker lets us package the app and its dependencies into a container that runs the same way on any machine. No "works on my machine" problems.

## What Docker actually does

If you've never used Docker, here's the short version: it bundles three Linux kernel features to isolate your app.

- **pivot_root** — filesystem isolation. Your app gets its own root filesystem, separate from the host.
- **Namespaces** — visibility isolation. Your app sees its own process tree, network stack, and hostname.
- **Cgroups** — resource limits. You can cap how much CPU and memory the container uses.

On top of these kernel primitives, Docker adds the developer tooling: images (filesystem snapshots), a build system (Dockerfile), a registry (Docker Hub), and a daemon that orchestrates it all.

## Choosing a base image

Docker images are built on top of other images. For a Go app, you start from an official Go image that has the compiler and toolchain pre-installed.

On [Docker Hub](https://hub.docker.com/_/golang), you'll see many tags. The ones that matter:

| Tag | What it is |
|-----|------------|
| `golang:1.25` | Go on Debian. The standard choice. |
| `golang:1.25-alpine` | Go on Alpine Linux. Much smaller (~250MB vs ~800MB). |
| `golang:1.25-bookworm` | Go on a specific Debian version. Only needed if you require a particular release. |

**Alpine** is a minimal Linux distribution. Smaller image means faster pulls and a smaller attack surface. The tradeoff is it uses `musl` instead of `glibc`, which can cause issues with some C dependencies. For a pure Go API like ours, Alpine works perfectly.

**Match your local Go version.** Check with `go version` and use the corresponding tag. If you're on Go 1.25, use `golang:1.25-alpine`.

## Writing the Dockerfile

A `Dockerfile` (capital D, no extension) is a recipe that builds an image layer by layer, from top to bottom.

```dockerfile
FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o go-todo-api
EXPOSE 8080
CMD ["./go-todo-api"]
```

Let's go through each instruction.

### FROM golang:1.25-alpine

Start from a base image that has Go installed. This is your foundation — like starting from a fresh machine that already has the right Go version.

### WORKDIR /app

Set the working directory inside the container. All subsequent commands run from here. This doesn't refer to a folder on your machine — it creates `/app` inside the container's filesystem.

Without a `WORKDIR`, your files would end up in `/` — the container's root — mixed in with system directories:

```
/
├── bin/          # system binaries
├── etc/          # configuration files
├── home/         # user home directories
├── usr/          # installed programs
├── var/          # logs, temp files
├── go.mod        # ← your files dumped here
├── go.sum
├── main.go
├── handlers.go
└── models.go
```

With `WORKDIR /app`, your files get their own clean directory:

```
/
├── bin/
├── etc/
├── home/
├── usr/
├── var/
└── app/          # ← your files live here
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── handlers.go
    └── models.go
```

### COPY go.mod go.sum ./

Copy the dependency files into the container. `COPY` takes one or more source files followed by a destination. When copying multiple files, the destination must be a directory (trailing slash):

```dockerfile
COPY go.mod go.sum ./          # multiple files → destination must end with /
COPY main.go ./main.go         # single file → can be a file path
COPY . ./                      # entire directory
```

The destination is relative to `WORKDIR`, so `./` means `/app/`.

We copy these **before** the source code — this is important for layer caching (explained below).

### RUN go mod download

Download the dependencies. Same as running this on your machine, but inside the container.

### COPY . ./

Now copy the rest of the source code.

### RUN go build -o go-todo-api

Compile the binary. The `-o` flag specifies the output name explicitly rather than relying on Go's default naming.

### EXPOSE 8080

Documents which port the app listens on. This is **metadata only** — it doesn't actually open the port. Think of it as a comment for anyone reading the Dockerfile.

### CMD ["./go-todo-api"]

The command to run when the container starts. This is what `docker run` executes.

## Layer caching: why the order matters

Each Dockerfile instruction creates a **layer**. Docker caches these layers and reuses them if the inputs haven't changed. But here's the key: **once a layer invalidates, every layer after it rebuilds too.**

This is why we copy `go.mod` and `go.sum` before the source code:

```
COPY go.mod go.sum ./    ← changes rarely
RUN go mod download      ← cached if go.mod didn't change
COPY . ./                ← changes often (any code edit)
RUN go build             ← rebuilds every time source changes
```

If we had copied everything at once (`COPY . ./`) and then ran `go mod download`, changing a single line in `handlers.go` would invalidate the copy layer, forcing a full dependency re-download on every build.

With the split approach, editing source code only invalidates from `COPY . ./` onwards — the dependency download is cached.

## Building the image

Make sure Docker Desktop is running, then:

```bash
docker build -t go-todo-api .
```

- `-t go-todo-api` tags the image with a name (otherwise you get an unreadable hash)
- `.` is the build context — the directory containing the Dockerfile and files to copy

## Running the container

```bash
docker run -p 8080:8080 go-todo-api
```

The `-p` flag maps a port on your machine to a port in the container:

```
-p <host-port>:<container-port>
```

Without `-p`, the container's port 8080 is isolated — nothing on your machine can reach it. Remember, `EXPOSE` in the Dockerfile is just documentation; `-p` is what actually opens the door.

## Try it out

With the container running:

```bash
curl http://localhost:8080/todos
```

Your API responds exactly as before — but now it's running inside a container, not directly on your machine.

## What's next

This is a basic single-stage Dockerfile. There's more to explore — multi-stage builds (for smaller production images), `.dockerignore` files, and `docker-compose` for running multiple containers together. We'll revisit these as the project grows.
