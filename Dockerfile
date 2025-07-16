

FROM golang:1.24 as builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd/bot
RUN go build -o /app/bot .

FROM debian:bookworm-slim

WORKDIR /app
COPY --from=builder /app/bot .

CMD ["./bot"]

