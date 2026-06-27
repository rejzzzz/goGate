FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o service-d .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/service-d .
EXPOSE 8083
CMD ["./service-d"]
