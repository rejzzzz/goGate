FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o service-a .

FROM scratch
WORKDIR /app
COPY --from=builder /build/service-a .
EXPOSE 8081
CMD ["./service-a"]
