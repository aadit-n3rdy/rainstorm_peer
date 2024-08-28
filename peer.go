package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/aadit-n3rdy/rainstorm/common"
	"strings"
	"net"

	"github.com/quic-go/quic-go"
)

func fetchFDD(fileID string, trackerIP string) (common.FileDownloadData, error) {
	dict := map[string]interface{} {
		"class": "init",
		"type": "download_start",
		"file_id": fileID,
	}
	dictMsg, err := json.Marshal(dict)

	conn, err := net.Dial("tcp", trackerIP+":"+fmt.Sprint(common.TRACKER_TCP_PORT))
	defer conn.Close()

	if err != nil {
		return common.FileDownloadData{}, err
	}

	conn.Write(dictMsg)

	buf := make([]byte, 1024)
	fdd := common.FileDownloadData{}
	n, err := conn.Read(buf)
	err = json.Unmarshal(buf[:n], &fdd)
	if err != nil {
		return common.FileDownloadData{}, err
	}
	return fdd, nil
}

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
