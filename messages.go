package main;

type FileReqMsg struct {
	FileName string `json:"file_name"`
	FileID string `json:"file_id"`
};

type ChunkAvailMsg struct {
	Status StatusCode `json:"status"`;
	Chunks []int `json:"chunks"`;
};

type ChunkReqMsg struct {
	Status StatusCode `json:"status"`;
	Chunk int `json:"chunk"`;
};

type FileTransferMsg struct {
	Status StatusCode `json:"status"`;
	Data string `json:"data"`; // Base 64 encoded
}
