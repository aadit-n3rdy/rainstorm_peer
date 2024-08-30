package main

import (
	"os"
	"sync"
	"errors"
	"fmt"
	"github.com/google/uuid"
);

const CHUNK_SIZE int64 = 1024;

type Chunk struct {
	FileName string;
	Done bool;
};

type Chunker struct {
	chunkCache map[uuid.UUID][]Chunk; // Map File ID to []Chunk
	chunkMutex sync.Mutex;
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
	self.chunkCache = make(map[uuid.UUID][]Chunk)
	return err
}

func (self *Chunker) getChunks(fileID uuid.UUID) ([]Chunk, error) {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()
	return self.getChunksUnsafe(fileID)
}

func (self *Chunker) getChunksUnsafe(fileID uuid.UUID) ([]Chunk, error) {
	dict, ok := self.chunkCache[fileID]
	if !ok {
		return []Chunk{}, errors.New(fmt.Sprintf("Unknown file ID %v", fileID))
	} 
	return dict, nil
}


func (self *Chunker) addDiskFile(fname string) (uuid.UUID, error) {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()

	fmt.Println("Got chunkmutex lock")

	// Chunks the file with the given file name, and returns a UUID to refer to it
	fileID := uuid.New()
	fmt.Println("Created fileID", fileID.String())

	fstat, err := os.Stat(fname)
	if err != nil {
		fmt.Println("huh, what file?")
		return fileID, err
	}
	size := fstat.Size()
	chunks := int(size / CHUNK_SIZE)
	if CHUNK_SIZE%1024 > 0 {
		chunks += 1
	}
	cf := make([]Chunk, chunks)
	fmt.Println("Created", chunks, "chunks.")
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
		cf[chunk] = Chunk{FileName: cfname, Done: true};
	}
	self.chunkCache[fileID] = cf
	return fileID, nil
}

func (self *Chunker) addEmptyFile(n_chunks int) uuid.UUID {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()

	fileID := uuid.New()

	arr := make([]Chunk, n_chunks)
	self.chunkCache[fileID] = arr

	hash := fileID.String()
	for i:= 0; i < n_chunks; i+=1 {
		cfname := fmt.Sprintf("%v/%v_%v.chunk", self.chunkPath, hash, i)
		arr[i] = Chunk{FileName: cfname, Done: false};
	}

	return fileID
}

func (self *Chunker) getChunkFname(fileID uuid.UUID, chunk int) (string, error) {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()

	dict, ok := self.chunkCache[fileID]
	if !ok {
		return "", errors.New(fmt.Sprintf("Unkown fileID %v", fileID.String()))
	}
	res := dict[chunk]
	return res.FileName, nil
}

func (self *Chunker) isChunkDone(fileID uuid.UUID, chunk int) bool {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()

	dict, ok := self.chunkCache[fileID]
	if !ok {
		return false
	}
	res := dict[chunk]
	if !res.Done {
		return false
	} else {
		return true
	}
}

func (self *Chunker) setChunkDone(fileID uuid.UUID, chunk int) error {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()

	dict, ok := self.chunkCache[fileID]
	if !ok {
		return errors.New(fmt.Sprintf("Unknown fileID %v", fileID.String()))
	}
	cur := dict[chunk]
	cur.Done = true
	dict[chunk] = cur
	return nil
}

func (self *Chunker) deleteFile(fileID uuid.UUID) {
	self.chunkMutex.Lock()
	defer self.chunkMutex.Unlock()

	delete(self.chunkCache, fileID)
}

func (self *Chunker) unchunk(fileID uuid.UUID, dest string) error {
	self.chunkMutex.Lock()
	fmt.Println("got lock")
	defer self.chunkMutex.Unlock()

	wf, err := os.Create(dest)
	defer wf.Close()
	if err != nil {
		return err
	}

	fmt.Println("Created dest", dest)

	cd, err := self.getChunksUnsafe(fileID)
	if err != nil {
		return err
	}
	
	chunks := len(cd)

	fmt.Println("Total of", chunks, " chunks")

	buf := make([]byte, 1024)
	for i := 0; i < chunks; i++ {
		fchunk := cd[i]
		if !fchunk.Done {
			return errors.New(fmt.Sprintf("Chunk %v is not done", i))
		}
		f, err := os.Open(fchunk.FileName)
		defer f.Close()
		if err != nil {
			return err
		}
		n, err := f.Read(buf)
		wf.Write(buf[:n])
	}
	return nil
}
