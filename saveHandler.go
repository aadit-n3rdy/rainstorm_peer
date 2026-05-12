package main

import "os"

func SaveAll(savepath string, chunker *Chunker) {
	os.MkdirAll(savepath, 0777)

	err := FileManagerSave(savepath + "/filemanager.csv")
	if err != nil {
		appLogger.Error().Err(err).Str("save_path", savepath).Msg("error saving file manager")
	}
	err = chunker.saveChunker()
	if err != nil {
		appLogger.Error().Err(err).Str("save_path", savepath).Msg("error saving chunker")
	}
	err = SaveReceivers(savepath + "/receivers.csv")
	if err != nil {
		appLogger.Error().Err(err).Str("save_path", savepath).Msg("error saving receivers")
	}
}

func LoadAll(savepath string, chunker *Chunker) {
	err := FileManagerLoad(savepath + "/filemanager.csv")
	if err != nil {
		appLogger.Error().Err(err).Str("save_path", savepath).Msg("error loading file manager")
	}
	err = chunker.loadChunker()
	if err != nil {
		appLogger.Error().Err(err).Str("save_path", savepath).Msg("error loading chunker")
	}
	err = LoadReceivers(savepath+"/receivers.csv", chunker)
	if err != nil {
		appLogger.Error().Err(err).Str("save_path", savepath).Msg("error loading receivers")
	}
}
