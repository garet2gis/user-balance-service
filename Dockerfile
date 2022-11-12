FROM golang:1.18 AS builder

COPY . /user_balance_service/
WORKDIR /user_balance_service/

RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.7
RUN make swagger
RUN CGO_ENABLED=0 GOOS=linux go build -o ./.bin/main ./cmd/main/

FROM alpine:3.15

WORKDIR /root/

COPY --from=builder /user_balance_service/.bin/main .
COPY --from=builder /user_balance_service/.env .
COPY --from=builder /user_balance_service/migrations ./migrations

ENTRYPOINT ["./main"]