package main

import (
	"os"
	"sync"
	"errors"
	"fmt"
	"hash/fnv"
);

const CHUNK_SIZE int64 = 1024;

type ChunkedFile struct {
	chunkData map[int]string
};

type Chunker struct {
	chunkCache sync.Map;
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

func (self *Chunker) chunkFile(fname string) (ChunkedFile, error) {
	rawVal, ok := self.chunkCache.Load(fname);
	if ok {
		fmt.Println("Chunk cache hit for ", fname)
		val, ok := rawVal.(ChunkedFile);
		if !ok {
			return ChunkedFile{}, errors.New("Invalid chunk cache entry for " + fname)
		}
		return val, nil
	}
	// No entry in chunkCache, 
	fstat, err := os.Stat(fname)
	if err != nil {
		return ChunkedFile{}, err
	}
	size := fstat.Size()
	chunks := int(size / CHUNK_SIZE)
	if CHUNK_SIZE%1024 > 0 {
		chunks += 1
	}
	var cf ChunkedFile
	cf.chunkData = make(map[int]string)
	f, err := os.Open(fname)
	if err != nil {
		return ChunkedFile{}, err
	}
	buf := make([]byte, CHUNK_SIZE)
	hasher := fnv.New32()
	hasher.Write([]byte(fname))
	hash := hasher.Sum32()
	for chunk := 0; chunk < chunks; chunk++ {
		n, err := f.Read(buf)
		if err != nil {
			return ChunkedFile{}, err
		}
		if n < int(CHUNK_SIZE) && chunk != chunks-1 {
			return ChunkedFile{}, errors.New("Unexpected read size drop")
		}

		cfname := fmt.Sprintf("%v/%v_%v.chunk", self.chunkPath, hash, chunk)
		fwrite, err := os.Create(cfname)
		defer fwrite.Close()
		if err != nil {
			return ChunkedFile{}, err
		}
		_, err = fwrite.Write(buf)
		if err != nil {
			return ChunkedFile{}, err
		}
		cf.chunkData[chunk] = cfname;
	}
	self.chunkCache.Store(fname, cf)

	return cf, nil
}
