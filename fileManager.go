package main

import (
	"sync"
	"github.com/google/uuid"
)

var FileManager sync.Map

type StoredFile struct {
	FileID string;
	FileName string;
	ChunkerID uuid.UUID;
}
