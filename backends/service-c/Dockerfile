FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o service-c .

FROM scratch
WORKDIR /app
COPY --from=builder /build/service-c .
EXPOSE 9000
CMD ["./service-c"]
