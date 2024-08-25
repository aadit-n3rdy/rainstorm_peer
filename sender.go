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
		//fmt.Println("File request msg: ", string(recvBuf[:n]))
		frm := FileReqMsg{}
		err = json.Unmarshal(recvBuf[:n], &frm)
		if err != nil {
			fmt.Printf("Couldn't unmarshall frm, %v\n", err)
			return
		}

		fmt.Println(frm)
		
		// 2. check file cache
		// NEED A MAPPING FROM FILEID to CHUNKFILEID
		// FN JUST SENDING THE SAME FILE
		rawsf, ok := FileManager.Load(frm.FileID)
		if !ok {
			stream.Write([]byte(fmt.Sprintf("{\"status\": %v}", STATUS_MISSING)))
			fmt.Printf("Unkown file ID %v\n", frm.FileID)
			return
		}
		sf := rawsf.(StoredFile)
		cd, err := chunker.getChunks(sf.ChunkerID)
		if err != nil {
			fmt.Printf("Unknown chunker id %v %v\n", sf.ChunkerID.String(), err)
			return
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

			fname, err := chunker.getChunkFname(sf.ChunkerID, crm.Chunk)
			if err != nil {
				fmt.Println(err)
				break
			}

			writeBuf := make([]byte, CHUNK_SIZE)
			f, err := os.Open(fname)
			n, err = f.Read(writeBuf)
			stream.Write(writeBuf[:n])
		}
		conn.CloseWithError(quic.ApplicationErrorCode(0), "bye!")
		fmt.Println("DONE")
	}
}


