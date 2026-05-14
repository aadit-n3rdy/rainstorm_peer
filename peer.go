package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	common "github.com/aadit-n3rdy/rainstorm_common"

	"github.com/quic-go/quic-go"
)

func PushHandler(local_fname string, fid string, fname string, trackerIP string, chunker *Chunker) error {
	chunkerID, err := chunker.addDiskFile(local_fname)
	if err != nil {
		appLogger.Error().Err(err).Str("file", local_fname).Msg("chunker failed to add disk file")
		return err
	}
	FileManagerAddFile(
		StoredFile{
			FileID:    fid,
			FileName:  fname,
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
		appLogger.Error().Err(err).Str("file_id", fid).Str("tracker_ip", trackerIP).Msg("failed to push file metadata")
		return err
	}
	appLogger.Info().Str("file", fname).Str("file_id", fid).Msg("successfully pushed file")
	return nil
}

func PullHandler(local_fname string, fid string, trackerIP string, chunker *Chunker, onComplete func(int, int, error)) {
	AddFileReceiver(fid, local_fname, trackerIP, chunker, onComplete)
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		appLogger.Error().Err(err).Msg("failed to determine working directory")
		return
	}
	opts, err := ParseRuntimeOptions(os.Args[1:], os.Getenv, wd)
	if err != nil {
		appLogger.Error().Err(err).Msg("failed to parse runtime options")
		return
	}
	if err := ConfigureLogger(opts.Config); err != nil {
		appLogger.Error().Err(err).Msg("failed to configure logger")
		return
	}

	chunker := &Chunker{}
	if err := chunker.init(opts.Config.ChunkPath); err != nil {
		appLogger.Error().Err(err).Str("chunk_path", opts.Config.ChunkPath).Msg("failed to initialize chunk storage")
		return
	}

	TrackerManagerInit()

	ReceiverInit()

	go aliveHandler()

	listener, err := quic.ListenAddr(fmt.Sprintf(
		":%v", common.PEER_QUIC_PORT),
		generateTLSConfig(),
		nil,
	)
	if err != nil {
		appLogger.Error().Err(err).Int("port", common.PEER_QUIC_PORT).Msg("could not listen for peer traffic")
		return
	}

	go sendHandler(listener, chunker)

	if opts.CLI {
		runCLI(chunker, opts.Config.SavePath)
	} else {
		StartGUI(chunker, opts.Config.SavePath)
	}
}

func runCLI(chunker *Chunker, SAVE_PATH string) {
	done := false
	var s string
	for !done {
		fmt.Scanln(&s)
		tokens := strings.Fields(s)
		if len(tokens) == 0 {
			continue
		}
		switch tokens[0] {
		case "push":
			fmt.Print("Enter local file name: ")
			var local_fname, fid, fname, trackerIP string
			fmt.Scanf("%s", &local_fname)
			fmt.Print("Enter file ID: ")
			fmt.Scanf("%s", &fid)
			fmt.Print("Enter file name: ")
			fmt.Scanf("%s", &fname)
			fmt.Print("Enter tracker IP: ")
			fmt.Scanf("%s", &trackerIP)
			PushHandler(local_fname, fid, fname, trackerIP, chunker)
		case "pull":
			fmt.Print("Enter local file name: ")
			var local_fname, fid, trackerIP string
			fmt.Scanf("%s", &local_fname)
			fmt.Print("Enter file ID: ")
			fmt.Scanf("%s", &fid)
			fmt.Print("Enter tracker IP: ")
			fmt.Scanf("%s", &trackerIP)
			PullHandler(local_fname, fid, trackerIP, chunker, nil)
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
}

func generateTLSConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		appLogger.Error().Err(err).Msg("failed to load TLS certificates")
	}
	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		NextProtos:         []string{"quic-rainstorm-p2p"},
	}
}
