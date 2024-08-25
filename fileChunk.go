package main

import (
	"os"
	"sync"
	"errors"
	"fmt"
	"github.com/google/uuid"
);

const CHUNK_SIZE int64 = 1024;

// type ChunkedFile struct {
// 	chunkData map[int]string
// };

type Chunker struct {
	chunkCache sync.Map; // Map File ID to ChunkedFile
	chunkPath string;
};

func (self *Chunker) init(chunkPath string) error {
	self.chunkPath = chunkPath;
	_, err := os.Stat(chunkPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(chunkPath, 0777);
	} else if err != nil {
		return err
	}
	return err
}

func (self *Chunker) getChunks(fileID uuid.UUID) (map[int]string, error) {
	anytype, ok := self.chunkCache.Load(fileID)
	if !ok {
		return map[int]string{}, errors.New(fmt.Sprintf("Unknown file ID %v", fileID))
	} 
	dict := anytype.(map[int]string)
	return dict, nil
}

func (self *Chunker) addChunkedFile(fname string) (uuid.UUID, error) {
	// Chunks the file with the given file name, and returns a UUID to refer to it
	fileID := uuid.New()
	_, err := self.getChunks(fileID)
	for err == nil {
		fileID = uuid.New()
	}

	fstat, err := os.Stat(fname)
	if err != nil {
		return fileID, err
	}
	size := fstat.Size()
	chunks := int(size / CHUNK_SIZE)
	if CHUNK_SIZE%1024 > 0 {
		chunks += 1
	}
	cf := make(map[int]string)
	f, err := os.Open(fname)
	if err != nil {
		return fileID, err
	}
	buf := make([]byte, CHUNK_SIZE)
	//hasher := fnv.New32()
	//hasher.Write([]byte(fileID.String()))
	//hash := hasher.Sum32()
	hash := fileID.String()
	for chunk := 0; chunk < chunks; chunk++ {
		n, err := f.Read(buf)
		if err != nil {
			return fileID, err
		}
		if n < int(CHUNK_SIZE) && chunk != chunks-1 {
			return fileID, errors.New("Unexpected read size drop")
		}

		cfname := fmt.Sprintf("%v/%v_%v.chunk", self.chunkPath, hash, chunk)
		fwrite, err := os.Create(cfname)
		defer fwrite.Close()
		if err != nil {
			return fileID, err
		}
		_, err = fwrite.Write(buf)
		if err != nil {
			return fileID, err
		}
		cf[chunk] = cfname;
	}
	self.chunkCache.Store(fileID, cf)

	return fileID, nil
}

func (self *Chunker) addEmptyFile() uuid.UUID {
	fileID := uuid.New()
	self.chunkCache.Store(fileID, map[int]string{})
	return fileID
}

func (self *Chunker) getChunkFname(fileID uuid.UUID, chunk int) (string, error) {
	res, ok := self.chunkCache.Load(fileID)
	if !ok {
		return "", errors.New(fmt.Sprintf("Unkown fileID %v", fileID.String()))
	}
	dict := res.(map[int]string)
	res, ok = dict[chunk]
	if ok {
		return res.(string), nil
	}
	fname := fmt.Sprintf("%v/%v_%v.chunk", self.chunkPath, fileID.String(), chunk)
	dict[chunk] = fname
	return fname, nil
}

func (self *Chunker) deleteFile(fileID uuid.UUID) {
	_, ok := self.chunkCache.Load(fileID)
	if ok {
		self.chunkCache.Delete(fileID)
	}
}

func (self *Chunker) unchunk(fileID uuid.UUID, dest string) error {
	wf, err := os.Create(dest)
	defer wf.Close()
	if err != nil {
		return err
	}

	cd, err := self.getChunks(fileID)
	if err != nil {
		return err
	}
	n_chunks := len(cd)
	buf := make([]byte, 1024)
	for i := 0; i < n_chunks; i++ {
		f, err := os.Open(cd[i])
		defer f.Close()
		if err != nil {
			return err
		}
		n, err := f.Read(buf)
		wf.Write(buf[:n])
	}
	return nil
}
