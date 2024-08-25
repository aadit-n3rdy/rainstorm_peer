package main

import (
	"testing"
	"fmt"
	"rainstorm/common"
	"github.com/quic-go/quic-go"
)

func TestSender(t *testing.T) {

	chunker := &Chunker{}
	chunker.init("./sender_chunk_path")
	var err error
	testFileID, err = chunker.addChunkedFile("./testfiles/file_100K.jpg")
	if err != nil {
		fmt.Println("Err: ", err)
		return
	}

	listener, err := quic.ListenAddr(fmt.Sprintf(
		":%v", common.PEER_QUIC_PORT),
		generateTLSConfig(),
		nil,
	)
	if err != nil {
		fmt.Println("Could not lisen on port ", common.PEER_QUIC_PORT, err);
		return
	}

	sendHandler(listener, chunker)
}
