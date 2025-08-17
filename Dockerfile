FROM golang:1.24-alpine AS go-builder
WORKDIR /src

COPY . .

RUN go mod download
RUN go build -ldflags="-s -w" -o /src/build/app ./cmd/app

FROM node:lts-alpine AS unocss-builder
WORKDIR /src

COPY ./web/ ./web/
COPY ./uno.config.js ./

RUN npm install unocss @unocss/preset-wind4 @unocss/reset
RUN npm exec unocss -- -c ./uno.config.js -o /src/build/uno.css --minify --preflights

FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=go-builder /src/build/app /app/app
COPY --from=go-builder /src/web/assets /app/web/assets
COPY --from=unocss-builder /src/build/uno.css /app/web/assets/css/uno.css

CMD ["./app"]
