package main

import (
	"fmt"
	"github.com/aadit-n3rdy/rainstorm/common"
	"testing"

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

	TrackerManagerInit()

	//fmt.Print("Enter tracker IP: ")
	//var trackerIP string 
	//fmt.Scanf("%s", &trackerIP)
	//fmt.Printf("Tracker is at: %v\n", trackerIP)
	trackerIP := "127.0.0.1"

	FileManagerAddFile(
		StoredFile{
			FileID: "somefileid", 
			FileName: "somefilename", 
			ChunkerID: chunkerID,
			TrackerIP: trackerIP,
		},
	)

	go aliveHandler()

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
