FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY . .

RUN go mod download
RUN go build -ldflags="-s -w" -o /app ./cmd/app

FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=builder /app /app
CMD ["/app"]
