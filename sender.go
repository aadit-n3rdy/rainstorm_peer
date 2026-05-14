package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	common "github.com/aadit-n3rdy/rainstorm_common"
	"github.com/google/uuid"

	"net"

	"github.com/quic-go/quic-go"
)

var testFileID uuid.UUID

func pushFDD(fdd *common.FileDownloadData, trackerIP string) error {

	smallfdd := common.FileDownloadData{
		FileID:     fdd.FileID,
		FileName:   fdd.FileName,
		Peers:      fdd.Peers,
		Checksums:  []string{},
		ChunkCount: fdd.ChunkCount,
	}

	dict := map[string]interface{}{
		"class":              "init",
		"type":               "file_register",
		"file_download_data": smallfdd,
	}

	conn, err := net.Dial("tcp", trackerIP+":"+fmt.Sprint(common.TRACKER_TCP_PORT))

	if err != nil {
		return err
	}
	defer conn.Close()

	fdd_msg, err := json.Marshal(dict)
	_, err = conn.Write(fdd_msg)
	if err != nil {
		return err
	}

	buf := []byte("ABCD")

	conn.Read(buf)
	for i := 0; i < fdd.ChunkCount; i += 1 {
		conn.Write([]byte(fmt.Sprintf("%v\n", fdd.Checksums[i])))
	}

	return nil
}

func sendHandler(listener *quic.Listener, chunker *Chunker) {
	for {
		// 0. listen for connections
		conn, err := listener.Accept(context.Background())
		if err != nil {
			appLogger.Error().Err(err).Msg("could not accept peer connection")
			return
		}
		go sendHandlerStream(conn, chunker)
	}
}

func sendHandlerStream(conn quic.Connection, chunker *Chunker) {
	defer conn.CloseWithError(quic.ApplicationErrorCode(0), "bye!")

	// 1. get file ID and file Name
	//fmt.Println("New connection from ", conn.RemoteAddr().String())
	stream, err := conn.OpenStream()
	if err != nil {
		appLogger.Warn().Err(err).Str("peer", conn.RemoteAddr().String()).Msg("could not open stream to peer")
		return
	}
	defer stream.Close()
	helloDict := map[string]interface{}{
		"status": STATUS_OK,
	}
	helloMsg, err := json.Marshal(helloDict)
	if err != nil {
		appLogger.Warn().Err(err).Str("peer", conn.RemoteAddr().String()).Msg("could not marshal hello message")
		return
	}
	n, err := stream.Write(helloMsg)
	if err != nil {
		appLogger.Warn().Err(err).Int("bytes_sent", n).Str("peer", conn.RemoteAddr().String()).Msg("could not send hello message")
		return
	}

	recvBuf := make([]byte, 1024)
	n, err = stream.Read(recvBuf)
	if err != nil {
		appLogger.Warn().Err(err).Str("peer", conn.RemoteAddr().String()).Msg("could not read file request message")
		return
	}
	//fmt.Println("File request msg: ", string(recvBuf[:n]))
	frm := FileReqMsg{}
	err = json.Unmarshal(recvBuf[:n], &frm)
	if err != nil {
		appLogger.Warn().Err(err).Str("peer", conn.RemoteAddr().String()).Msg("could not unmarshal file request message")
		return
	}

	// 2. check file cache
	sf, ok := FileManagerGetFile(frm.FileID)
	if !ok {
		stream.Write([]byte(fmt.Sprintf("{\"status\": %v}", STATUS_MISSING)))
		appLogger.Warn().Str("file_id", frm.FileID).Str("peer", conn.RemoteAddr().String()).Msg("unknown requested file")
		return
	}
	cd, err := chunker.getChunks(sf.ChunkerID)
	if err != nil {
		appLogger.Warn().Err(err).Str("chunker_id", sf.ChunkerID.String()).Msg("unknown chunker id")
		return
	}

	cam := ChunkAvailMsg{}
	cam.Chunks = make([]int, len(cd))
	i := 0
	for k := range cd {
		cam.Chunks[i] = k
		i += 1
	}
	camBuf, err := json.Marshal(cam)
	if err != nil {
		appLogger.Warn().Err(err).Msg("could not marshal chunk availability message")
		return
	}
	n, err = stream.Write(camBuf)

	for true {
		n, err = stream.Read(recvBuf)
		if n == 0 {
			break
		}
		crm := ChunkReqMsg{}
		err = json.Unmarshal(recvBuf[:n], &crm)
		//fmt.Println(crm.Chunk)
		if crm.Status == STATUS_DONE {
			break
		}

		fname, err := chunker.getChunkFname(sf.ChunkerID, crm.Chunk)
		if err != nil {
			appLogger.Warn().Err(err).Int("chunk", crm.Chunk).Msg("could not resolve chunk filename")
			break
		}

		st, err := os.Stat(fname)
		if err != nil {
			appLogger.Warn().Err(err).Str("chunk_file", fname).Msg("could not stat chunk file")
			break
		}
		size := st.Size()

		var done int64 = 0

		writeBuf := make([]byte, 1024)
		f, err := os.Open(fname)

		size_bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(size_bytes, uint64(size))

		stream.Write(size_bytes)
		for done < size {
			n, err = f.Read(writeBuf)
			if n == 0 || err != nil {
				break
			}
			done += int64(n)
			stream.Write(writeBuf[:n])
		}
	}
}
