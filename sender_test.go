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
	chunkerID, err := chunker.addChunkedFile("./testfiles/file_100K.jpg")
	fmt.Printf("Chunked to: %v\n", chunkerID)
	if err != nil {
		fmt.Println("Err: ", err)
		return
	}

	FileManager.Store(
		"somefileid", 
		StoredFile{
			FileID: "somefileid", 
			FileName: "somefilename", 
			ChunkerID: chunkerID,
		},
	)

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
