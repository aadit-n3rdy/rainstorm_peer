package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"bufio"
	common "github.com/aadit-n3rdy/rainstorm_common"
	"time"
	"github.com/google/uuid"

	"github.com/quic-go/quic-go"
	"net"
	"sync"
);

const (
	RECV_DONE = iota
	RECV_FAIL
)

var PeerIPBlackList map[string]interface{}
var PeerBlackListMut sync.Mutex

func IsPeerBlackListed(peerIP string) bool {
	PeerBlackListMut.Lock()
	_, isblack := PeerIPBlackList[peerIP]
	PeerBlackListMut.Unlock()
	return isblack
}

func AddPeerToBlackList(peerIP string) {
	PeerBlackListMut.Lock()
	defer PeerBlackListMut.Unlock()
	PeerIPBlackList[peerIP] = struct{}{}
}

const MAX_RECV_THREADS = 10

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

	conn.Write([]byte("OK"))

	br := bufio.NewReader(conn)

	fdd.Checksums = make([]string, fdd.ChunkCount)

	for i := 0; i < fdd.ChunkCount; i+=1 {
		fdd.Checksums[i], _ = br.ReadString('\n')
		fdd.Checksums[i] = fdd.Checksums[i][:len(fdd.Checksums[i])-1]
	}

	return fdd, nil
}

func fileReceiver(fdd common.FileDownloadData, dest string, chunker *Chunker) error {
	if len(fdd.Peers) == 0 {
		return errors.New("No peers for given file")
	}
	trigChan := make(chan int)

	chunkerID := chunker.addEmptyFile(fdd.ChunkCount)

	next_peer := 0
	n_fail := 0
	n_peers := 0

	for n_peers < MAX_RECV_THREADS && next_peer < len(fdd.Peers) {
		go fileReceiveStream(fdd,chunkerID, fdd.Peers[next_peer], chunker, trigChan)
		n_peers += 1
		next_peer += 1
	}

	for {
		code := <- trigChan
		if code == RECV_DONE {
			break
		} else if code == RECV_FAIL {
			n_fail += 1
			n_peers -= 1
			if next_peer < len(fdd.Peers) {
				go fileReceiveStream(fdd, chunkerID, fdd.Peers[next_peer], chunker, trigChan)
				n_peers += 1
				next_peer += 1
			} else if n_fail == len(fdd.Peers) {
				// All peers have failed, return error
				if chunker.isFileDone(chunkerID)  {
					break
				}
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

		if IsPeerBlackListed(peer.IP) {
			trig <- RECV_FAIL
			return errors.New("Peer IP " + peer.IP + " was blacklisted")
		}

		destStr := fmt.Sprintf("%v:%v", peer.IP, peer.Port)
		conn, err := quic.DialAddr(context.Background(), destStr, generateTLSConfig(), &quicConf)
		if err != nil {
			trig <- RECV_FAIL
			return errors.New(fmt.Sprintf("Error dialing addr %v, %v", destStr, err))
		}

		fmt.Println("Dialed addr")

		stream, err := conn.AcceptStream(conn.Context())
		if err != nil {
			trig <- RECV_FAIL
			return errors.New(fmt.Sprintf("Error accepting stream from %v, %v", destStr, err))
		}

		fmt.Println("Accepted stream")

		buf := make([]byte, 1024)

		n, err := stream.Read(buf)

		dict := map[string]interface{}{}

		err = json.Unmarshal(buf[:n], &dict)
		if err != nil {
			trig <- RECV_FAIL
			return errors.New(fmt.Sprintf("Error unmarshalling %v hello, %v", string(buf[:n]), err))
		}
		fmt.Println(dict)

		frm := FileReqMsg{FileID: fdd.FileID, FileName: fdd.FileName}
		frmBuf, err := json.Marshal(frm)
		if err != nil {
			trig <- RECV_FAIL
			return errors.New(fmt.Sprintf("Error Marshaling FRM, %v", err))
		}

		n, err = stream.Write(frmBuf)
		if err != nil {
			trig <- RECV_FAIL
			return errors.New(fmt.Sprintf("Error writing FRM, %v", err))
		}

		n, err = stream.Read(buf)
		if err != nil {
			trig <- RECV_FAIL
			return err
		}

		iters := 0

		for (!chunker.isFileDone(chunkerID) && iters < 5) {
			cam := ChunkAvailMsg{}
			err = json.Unmarshal(buf[:n], &cam)
			fmt.Println(string(buf[:n]))
			if err != nil {
				trig <- RECV_FAIL
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
					trig <- RECV_FAIL
					return errors.New(fmt.Sprintf("Error marshaling CRM, %v", err))
				}
				_, err = stream.Write(crmBuf)
				if err != nil {
					trig <- RECV_FAIL
					return errors.New(fmt.Sprintf("Error writing CRM, %v", err))
				}

				fname, err := chunker.getChunkFname(chunkerID, cam.Chunks[i])
				if err != nil {
					trig <- RECV_FAIL
					return errors.New(fmt.Sprintf("Error getting chunk %v fname, %v", cam.Chunks[i], err))
				}
				chunker.markChunkDone(chunkerID, cam.Chunks[i])
				f, err := os.Create(fname)
				if err != nil {
					trig <- RECV_FAIL
					return errors.New("Could not open file: " + fname + " " + err.Error())
				}

				n, err = stream.Read(buf)
				f.Write(buf[:n])
				f.Close()

				verified, _ :=  chunker.verifyChunk(chunkerID, cam.Chunks[i], fdd.Checksums[cam.Chunks[i]])
				if !verified {
					fmt.Printf("Chunk %d had hash %v failed\n", cam.Chunks[i], fdd.Checksums[cam.Chunks[i]])
					chunker.deleteChunk(chunkerID, cam.Chunks[i])
					AddPeerToBlackList(peer.IP)
					trig <- RECV_FAIL
					return errors.New(fmt.Sprintf("Checksum fail on chunk %v", cam.Chunks[i]))
				}

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
		return nil
	}
