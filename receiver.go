package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"rainstorm/common"
	"time"

	"github.com/quic-go/quic-go"
);

func fileReceiver(fdd common.FileDownloadData, dest string, chunker *Chunker) error {
	quicConf := quic.Config{
	MaxIdleTimeout: 60 * time.Second,
	}
	destStr := fmt.Sprintf("%v:%v", fdd.Peers[0].IP, fdd.Peers[0].Port)
	conn, err := quic.DialAddr(context.Background(), destStr, generateTLSConfig(), &quicConf)
	if err != nil {
		return errors.New(fmt.Sprintf("Error dialing addr %v, %v", destStr, err))
	}

	fmt.Println("Dialed addr")

	chunkerID := chunker.addEmptyFile()

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
	cam := ChunkAvailMsg{}
	err = json.Unmarshal(buf[:n], &cam)
	fmt.Println(string(buf[:n]))
	if err != nil {
		return errors.New(fmt.Sprintf("Error Unmarshaling CAM, %v", err))
	}

	fmt.Println("Chunks available: ", cam.Chunks)

	for i:= 0; i < len(cam.Chunks); i++ {
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
		f, err := os.Create(fname)
		if err != nil {
			return errors.New("Could not open file: " + fname + " " + err.Error())
		}

		n, err = stream.Read(buf)
		f.Write(buf[:n])
		f.Close()
	}

	crm := ChunkReqMsg{Chunk: -1, Status: STATUS_DONE}
	crmBuf, err := json.Marshal(crm)
	stream.Write(crmBuf)

	chunker.unchunk(chunkerID, dest)

	fmt.Println("DONE!")
	stream.Close()

	return nil
}
