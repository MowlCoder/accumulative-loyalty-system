FROM golang:1.20-alpine as builder

WORKDIR /app
COPY . /app

RUN go clean --modcache
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o accrual ./cmd/accrual/main.go

FROM golang:alpine

COPY --from=builder /app/accrual .

EXPOSE 8081

CMD ["./accrual"]