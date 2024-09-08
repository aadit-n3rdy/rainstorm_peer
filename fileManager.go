package main

import (
	"sync"

	common "github.com/aadit-n3rdy/rainstorm_common"
	"github.com/google/uuid"
)

var FileManager sync.Map // Map from FileID to StoredFile

type StoredFile struct {
	FileID string;
	FileName string;
	ChunkerID uuid.UUID;
	TrackerIP string;
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
