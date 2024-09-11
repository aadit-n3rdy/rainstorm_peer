package main

import (
	"fmt"
	"os"
)


func SaveAll(savepath string, chunker *Chunker) {

	os.MkdirAll(savepath, 0777)

	err := FileManagerSave(savepath + "/filemanager.csv")
	if err != nil {
		fmt.Printf("Error saving FM: %v\n", err)
	}
	err = chunker.saveChunker()
	if err != nil {
		fmt.Printf("Error saving Chunker: %v\n", err)
	}
	err = SaveReceivers(savepath+"/receivers.csv")
	if err != nil {
		fmt.Printf("Error saving REC: %v\n", err)
	}
}

func LoadAll(savepath string, chunker *Chunker) {
	err := FileManagerLoad(savepath + "/filemanager.csv")
	if err != nil {
		fmt.Printf("Error loading FM: %v\n", err)
	}
	err = chunker.loadChunker()
	if err != nil {
		fmt.Printf("Error loading chunker: %v\n", err)
	}
	err = LoadReceivers(savepath+"/receivers.csv", chunker)
	if err != nil {
		fmt.Printf("Error loading REC: %v\n", err)
	}

}
