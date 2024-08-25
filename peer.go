package main

import (
	"fmt"
	"strings"
	"rainstorm/common"
	"github.com/quic-go/quic-go"
	"crypto/tls"
)

func main() {
	done := false
	var s string;

	chunker := &Chunker{}
	chunker.init("./chunkdata")

	listener, err := quic.ListenAddr(fmt.Sprintf(
		":%v", common.PEER_QUIC_PORT),
		generateTLSConfig(),
		nil,
	)
	if err != nil {
		fmt.Println("Could not lisen on port ", common.PEER_QUIC_PORT, err);
		return
	}

	go sendHandler(listener, chunker)

	for !done {
		fmt.Scanln(&s);
		tokens := strings.Fields(s)
		if len(tokens) == 0 {
			continue;
		}
	}
	return;
}

func generateTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		fmt.Printf("Failed to load TLS certificates: %v\n", err)
	}
	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-rainstorm-p2p"},
	}
}
