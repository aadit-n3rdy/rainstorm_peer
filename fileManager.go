package main

import (
	"io"
	"encoding/csv"
	"os"
	"sync"

	common "github.com/aadit-n3rdy/rainstorm_common"
	"github.com/google/uuid"
)

var FileManager sync.Map // Map from FileID to StoredFile

type StoredFile struct {
	FileID string `json:"file_id"`
	FileName string `json:"file_name"`
	ChunkerID uuid.UUID `json:"chunker_id"`
	TrackerIP string `json:"tracker"`
}

func FileManagerAddFile(sf StoredFile) {
	FileManager.Store(sf.FileID, sf)
	AddTracker(&sf)
}

func FileManagerRemoveFile(fileID string) {
	raw, ok := FileManager.Load(fileID)
	if !ok {
		return
	}
	sf := raw.(StoredFile)
	RemoveTracker(&sf)
}

func FileManagerGetFile(fileID string) (StoredFile, bool) {
	raw, ok := FileManager.Load(fileID)
	if !ok {
		return StoredFile{}, ok
	}
	return raw.(StoredFile), ok
}

func FileManagerFillFDD(fileID string, chunker *Chunker, fdd *common.FileDownloadData) {
	raw, ok := FileManager.Load(fileID)
	if !ok {
		return
	}
	sf := raw.(StoredFile)
	if !chunker.isFileDone(sf.ChunkerID) {
		return
	}
	fdd.FileID = fileID
	fdd.FileName = sf.FileName
	fdd.Peers = make([]common.Peer, 0)
	fdd.Checksums = chunker.getCheckSums(sf.ChunkerID)
	fdd.ChunkCount = len(fdd.Checksums)
}

func FileManagerSave(savefile string) error  {
	f, err := os.Create(savefile)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	FileManager.Range(
		func(key, value any) bool {
			v := value.(StoredFile)
			w.Write([]string{v.FileID, v.FileName, v.ChunkerID.String(), v.TrackerIP})
			return true
		},
	)


	return nil
}

func FileManagerLoad(savefile string) error {
	f, err := os.Open(savefile)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)

	rec, err := r.Read()

	for err != io.EOF {
		if len(rec) == 0 {
			rec, err = r.Read()
			continue
		}
		if err != nil {
			return err
		}

	 	cid, err := uuid.Parse(rec[2])
		if err != nil {
			return err
		}

		sf := StoredFile{
			FileID: rec[0],
			FileName: rec[1],
			ChunkerID: cid,
			TrackerIP: rec[3],
		}

		FileManagerAddFile(sf)
		rec, err = r.Read()
	}
	return nil
}
