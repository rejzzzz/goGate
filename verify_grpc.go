package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

func main() {
	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			// For h2c (cleartext HTTP/2) over standard net.Conn:
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	
	msg := []byte(`{"message": "hello proxy"}`)
	
	// Create gRPC frame: 1 byte compress flag (0), 4 bytes length
	frame := make([]byte, 5+len(msg))
	frame[0] = 0 // uncompressed
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(msg)))
	copy(frame[5:], msg)

	req, err := http.NewRequest("POST", "http://localhost:8080/echo.EchoService/Echo", bytes.NewReader(frame))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/grpc")
	
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	fmt.Printf("Status: %s\n", resp.Status)
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 5 {
		fmt.Printf("Proxy Success! Raw Response (unframed): %s\n", string(body[5:]))
	} else {
		fmt.Printf("Raw response bytes: %x\n", body)
	}
}
