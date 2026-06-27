FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o service-a .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/service-a .
EXPOSE 8081
CMD ["./service-a"]
