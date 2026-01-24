FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/facilitator ./cmd/facilitator

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/facilitator /usr/local/bin/facilitator

EXPOSE 4020

ENTRYPOINT ["facilitator"]
