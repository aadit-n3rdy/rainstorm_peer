package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	common "github.com/aadit-n3rdy/rainstorm_common"
	"time"
	"github.com/google/uuid"

	"github.com/quic-go/quic-go"
);

const (
	RECV_DONE = iota
	RECV_FAIL
)

const MAX_RECV_THREADS = 10

func fileReceiver(fdd common.FileDownloadData, dest string, chunker *Chunker) error {
	if len(fdd.Peers) == 0 {
		return errors.New("No peers for given file")
	}
	trigChan := make(chan int)

	chunkerID := chunker.addEmptyFile(fdd.ChunkCount)

	next_peer := 0
	n_fail := 0

	for next_peer < MAX_RECV_THREADS && next_peer < len(fdd.Peers) {
		go fileReceiveStream(fdd,chunkerID, fdd.Peers[next_peer], chunker, trigChan)
		next_peer += 1
	}

	for {
		code := <- trigChan
		if code == RECV_DONE {
			break
		} else if code == RECV_FAIL {
			n_fail += 1
			if next_peer < len(fdd.Peers) {
				go fileReceiveStream(fdd, chunkerID, fdd.Peers[next_peer], chunker, trigChan)
				next_peer += 1
			} else if n_fail == len(fdd.Peers) {
				// All peers have failed, return error
				return errors.New("All peers have failed")
			}
		} else {
			fmt.Println("Unknown code ", code)
		}
	}

	err := chunker.unchunk(chunkerID, dest)
	if  err != nil {
		fmt.Println("Failed while unchunking:", err)
	}
	fmt.Println("DONE UNCHUNKING!")
	return nil
}

func fileReceiveStream(
	fdd common.FileDownloadData, 
	chunkerID uuid.UUID, 
	peer common.Peer, 
	chunker *Chunker,
	trig chan int) error {
	quicConf := quic.Config{
		MaxIdleTimeout: 60 * time.Second,
	}
	destStr := fmt.Sprintf("%v:%v", peer.IP, peer.Port)
	conn, err := quic.DialAddr(context.Background(), destStr, generateTLSConfig(), &quicConf)
	if err != nil {
		return errors.New(fmt.Sprintf("Error dialing addr %v, %v", destStr, err))
	}

	fmt.Println("Dialed addr")


	stream, err := conn.AcceptStream(conn.Context())
	if err != nil {
		return errors.New(fmt.Sprintf("Error accepting stream from %v, %v", destStr, err))
	}

	fmt.Println("Accepted stream")

	buf := make([]byte, 1024)

	n, err := stream.Read(buf)

	dict := map[string]interface{}{}

	err = json.Unmarshal(buf[:n], &dict)
	if err != nil {
		return errors.New(fmt.Sprintf("Error unmarshalling %v hello, %v", string(buf[:n]), err))
	}
	fmt.Println(dict)

	frm := FileReqMsg{FileID: fdd.FileID, FileName: fdd.FileName}
	frmBuf, err := json.Marshal(frm)
	if err != nil {
		return errors.New(fmt.Sprintf("Error Marshaling FRM, %v", err))
	}

	n, err = stream.Write(frmBuf)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing FRM, %v", err))
	}

	n, err = stream.Read(buf)
	if err != nil {
		return err
	}

	iters := 0

	for (!chunker.isFileDone(chunkerID) && iters < 5) {
		cam := ChunkAvailMsg{}
		err = json.Unmarshal(buf[:n], &cam)
		fmt.Println(string(buf[:n]))
		if err != nil {
			return errors.New(fmt.Sprintf("Error Unmarshaling CAM, %v", err))
		}

		fmt.Println("Chunks available: ", cam.Chunks)

		for i:= 0; i < len(cam.Chunks); i++ {
			if (!chunker.markChunkBusyIfFree(chunkerID, cam.Chunks[i])) {
				continue;
			}
			crm := ChunkReqMsg{Chunk: cam.Chunks[i], Status: STATUS_OK}
			crmBuf, err := json.Marshal(crm)
			if err != nil {
				return errors.New(fmt.Sprintf("Error marshaling CRM, %v", err))
			}
			_, err = stream.Write(crmBuf)
			if err != nil {
				return errors.New(fmt.Sprintf("Error writing CRM, %v", err))
			}

			fname, err := chunker.getChunkFname(chunkerID, cam.Chunks[i])
			if err != nil {
				return errors.New(fmt.Sprintf("Error getting chunk %v fname, %v", cam.Chunks[i], err))
			}
			chunker.markChunkDone(chunkerID, cam.Chunks[i])
			f, err := os.Create(fname)
			if err != nil {
				return errors.New("Could not open file: " + fname + " " + err.Error())
			}

			n, err = stream.Read(buf)
			f.Write(buf[:n])
			f.Close()
		}
		iters += 1
	}
	crm := ChunkReqMsg{Chunk: -1, Status: STATUS_DONE}
	crmBuf, err := json.Marshal(crm)
	stream.Write(crmBuf)
	stream.Close()

	if chunker.isFileDone(chunkerID) {
		trig <- RECV_DONE
	} else {
		trig <- RECV_FAIL
	}

	fmt.Println("DONE!")

	return nil

}
