FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o service-b .

FROM scratch
WORKDIR /app
COPY --from=builder /build/service-b .
EXPOSE 8082
CMD ["./service-b"]
