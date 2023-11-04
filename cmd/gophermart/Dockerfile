FROM golang:1.20-alpine as builder

WORKDIR /app
COPY . /app

RUN go clean --modcache
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gophermart ./cmd/gophermart/main.go

FROM golang:alpine

COPY --from=builder /app/gophermart .

EXPOSE 8080

CMD ["./gophermart"]