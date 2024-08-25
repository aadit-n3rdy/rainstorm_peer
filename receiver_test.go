package main

import (
	"fmt"
	"testing"
	//"fmt"
	"rainstorm/common"
)

func TestReceiver(t *testing.T) {

	chunker := &Chunker{}
	chunker.init("./receiver_chunk_path")

	fmt.Println("Chunker ready")

	err := fileReceiver(
		common.FileDownloadData{
			FileID: "somefileid", 
			FileName: "somefilename", 
			Peers: []common.Peer{{IP: "127.0.0.1", Port: common.PEER_QUIC_PORT}},
		},
		"testfile.jpeg",
		chunker,
	)

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Test done!")
}
