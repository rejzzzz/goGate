FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod init service-d && CGO_ENABLED=0 go build -ldflags="-s -w" -o service-d .

FROM scratch
WORKDIR /app
COPY --from=builder /build/service-d .
EXPOSE 8083
CMD ["./service-d"]
