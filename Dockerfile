FROM golang:1.13 as builder

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://proxy.golang.org go build -o app ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8000

ENTRYPOINT ["./app"]
