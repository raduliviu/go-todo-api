FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o go-todo-api 
EXPOSE 8080
CMD ["./go-todo-api"]