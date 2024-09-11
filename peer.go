package main

import (
	"crypto/tls"
	"fmt"
	common "github.com/aadit-n3rdy/rainstorm_common"
	"strings"
	"os"

	"github.com/quic-go/quic-go"
)

func pushHandler(local_fname string , fid string, fname string, trackerIP string, chunker *Chunker) {
	chunkerID, err := chunker.addDiskFile(local_fname)
	if err != nil {
		fmt.Printf("Chunker error: %s\n", err.Error())
		return
	}
	FileManagerAddFile(
		StoredFile{
			FileID: fid,
			FileName: fname,
			ChunkerID: chunkerID,
			TrackerIP: trackerIP,
		},
	)
	var fdd common.FileDownloadData
	FileManagerFillFDD(
		fid,
		chunker,
		&fdd,
	)
	err = pushFDD(&fdd, trackerIP)
	if err != nil {
		fmt.Printf("Error pushing FDD: %s\n", err.Error())
		return
	}
}

func pullHandler(local_fname string, fid string, trackerIP string, chunker *Chunker) {
	AddFileReceiver(fid, local_fname, trackerIP, chunker)
}

func main() {
	done := false
	var s string;

	chunker := &Chunker{}
	SAVE_PATH := os.Getenv("RSTM_SAVE_PATH")
	if SAVE_PATH == "" {
		SAVE_PATH, _ = os.Getwd()
		SAVE_PATH = SAVE_PATH + "/rstm_save"
	}
	chunker.init(SAVE_PATH + "/chunk_path")

	TrackerManagerInit()

	ReceiverInit()

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

	go sendHandler(listener, chunker)

	for !done {
		fmt.Scanln(&s);
		tokens := strings.Fields(s)
		if len(tokens) == 0 {
			continue;
		}
		switch tokens[0] {
			case "push" :
				fmt.Print("Enter local file name: ")
				var local_fname, fid, fname, trackerIP string
				fmt.Scanf("%s", &local_fname)
				fmt.Print("Enter file ID: ")
				fmt.Scanf("%s", &fid)
				fmt.Print("Enter file name: ")
				fmt.Scanf("%s", &fname)
				fmt.Print("Enter tracker IP: ")
				fmt.Scanf("%s", &trackerIP)
				pushHandler(local_fname, fid, fname, trackerIP, chunker)
			case "pull":
				fmt.Print("Enter local file name: ")
				var local_fname, fid, trackerIP string
				fmt.Scanf("%s", &local_fname)
				fmt.Print("Enter file ID: ")
				fmt.Scanf("%s", &fid)
				fmt.Print("Enter tracker IP: ")
				fmt.Scanf("%s", &trackerIP)
				pullHandler(local_fname, fid, trackerIP, chunker)
			case "load":
				LoadAll(SAVE_PATH, chunker)
			case "save":
				SaveAll(SAVE_PATH, chunker)
			case "exit":
				fmt.Print("Saving and exiting...\n")
				SaveAll(SAVE_PATH, chunker)
				done = true
		}
	}
	return;
}

func generateTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		fmt.Printf("Failed to load TLS certificates: %v\n", err)
	}
	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-rainstorm-p2p"},
	}
}
