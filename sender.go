package main

import (
	"context"
	"encoding/json"
//	"encoding/base64"
	"fmt"
	"os"
//	"rainstorm/common"

	"github.com/google/uuid"

	"github.com/quic-go/quic-go"
);

var testFileID uuid.UUID;

const SUBCHUNK_SIZE int64 = 100;

func sendHandler(listener *quic.Listener, chunker *Chunker) {
	for {
		// 0. listen for connections
		conn, err := listener.Accept(context.Background())
		defer conn.CloseWithError(quic.ApplicationErrorCode(0), "bye!")
		if err != nil {
			fmt.Println("Could not accept conn from ", conn.RemoteAddr().String(), err);
			return
		}

		// 1. get file ID and file Name
		fmt.Println("New connection from ", conn.RemoteAddr().String())
		stream, err := conn.OpenStream()
		if err != nil {
			fmt.Println("Could not open stream to ", conn.RemoteAddr().String(), err);
			return
		}
		defer stream.Close()
		helloDict := map[string]interface{}{
			"status": STATUS_OK,
		};
		helloMsg, err := json.Marshal(helloDict)
		if err != nil {
			fmt.Println("Could not marshal hello msg ", conn.RemoteAddr().String(), err);
			return
		}
		n, err := stream.Write(helloMsg)
		if err != nil {
			fmt.Printf("Could not send hello\n%v bytes sent\nError: %v\n", n, err);
			return
		}

		recvBuf := make([]byte, 1024)
		n, err = stream.Read(recvBuf)
		if err != nil {
			fmt.Printf("Could not read frm from %v\nError: %v\n", conn.RemoteAddr().String(), err);
			return
		}
		fmt.Println("File request msg: ", string(recvBuf[:n]))
		//ackDict := map[string]interface{}{
		//	"status": STATUS_OK,
		//};

		// 2. check file cache
		// NEED A MAPPING FROM FILEID to CHUNKFILEID
		// TODO: FN JUST SENDING THE SAME FILE
		cd, err := chunker.getChunks(testFileID)
		if err != nil {
			fmt.Println(err)
		}

		cam := ChunkAvailMsg{}
		cam.Chunks = make([]int, len(cd))
		i := 0
		for k := range(cd) {
			cam.Chunks[i] = k
			i +=1
		}
		camBuf, err := json.Marshal(cam)
		if err != nil {
			fmt.Println("Couldnt marshal cam: ", err)
			return
		}
		n, err = stream.Write(camBuf)

		for true {
			n, err = stream.Read(recvBuf)
			crm := ChunkReqMsg{}
			err = json.Unmarshal(recvBuf[:n], &crm)
			fmt.Println(crm.Chunk)
			if crm.Status == STATUS_DONE {
				break
			}

			fname, err := chunker.getChunkFname(testFileID, crm.Chunk)
			if err != nil {
				fmt.Println(err)
				break
			}

			writeBuf := make([]byte, CHUNK_SIZE)
			f, err := os.Open(fname)
			n, err = f.Read(writeBuf)
			stream.Write(writeBuf[:n])

			//writeBuf := make([]byte, SUBCHUNK_SIZE)
			//f, err := os.Open(fname)
			//base64buf := make([]byte, (SUBCHUNK_SIZE+1)*4)
			//for true {
			//	n, err := f.Read(writeBuf)
			//	if err != nil {
			//		fmt.Println(err)
			//	}
			//	if n == 0 {
			//		break
			//	}
			//	base64.RawStdEncoding.Encode(base64buf, writeBuf)

			//	ftm := FileTransferMsg{Status: STATUS_OK, Data: string(base64buf)}
			//	ftmBuf, _ := json.Marshal(ftm)
			//	stream.Write(ftmBuf)
			//}
		}
		conn.CloseWithError(quic.ApplicationErrorCode(0), "bye!")
		fmt.Println("DONE")
	}
}


