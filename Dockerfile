# Multi-stage build
FROM golang:1.24-bullseye AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

# install templ tool
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

# generate templ Go files (if you use templ)
RUN go tool templ generate ./webstatic/templ || true

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app ./cmd/web

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app /app
ENV PORT=8080
EXPOSE 8080
CMD ["/app"]
