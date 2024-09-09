package main

import (
	"fmt"
)


func SaveAll(savepath string, chunker *Chunker) {
	err := FileManagerSave(savepath + "/filemanager.csv")
	if err != nil {
		fmt.Printf("Error saving FM: %v\n", err)
	}
	err = chunker.saveChunker()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Error saving Chunker: %v\n", err)
	}
}

func LoadAll(savepath string, chunker *Chunker) {
	err := FileManagerLoad(savepath + "/filemanager.csv")
	if err != nil {
		fmt.Printf("Error loading FM: %v\n", err)
	}
	err = chunker.loadChunker()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Error loading chunker: %v\n", err)
	}
}
