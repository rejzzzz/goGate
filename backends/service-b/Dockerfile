FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o service-b .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/service-b .
EXPOSE 8082
CMD ["./service-b"]
