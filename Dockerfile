FROM golang:1.21 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && go build -o cekkuota

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/cekkuota .

ENV PORT=3000

CMD ["./cekkuota"]