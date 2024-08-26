package main

import (
	"sync"
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
