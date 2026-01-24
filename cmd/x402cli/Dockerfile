FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/x402cli ./cmd/x402cli

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/x402cli /usr/local/bin/x402cli

ENTRYPOINT ["x402cli"]
