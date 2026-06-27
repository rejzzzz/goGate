FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o service-c .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/service-c .
EXPOSE 9000
CMD ["./service-c"]
