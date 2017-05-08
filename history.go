package main

import (
	"encoding/json"
	"os"
	"path"
)

type GrueHistory struct {
	path  string
	Feeds map[string]*RSSFeed
}

func (hist *GrueHistory) String() string {
	b, err := json.Marshal(hist)
	if err != nil {
		panic("Cant Marshal GrueHistory")
	}
	return string(b)
}

func (hist *GrueHistory) Write() error {
	file, err := os.Create(hist.path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(hist)
}

func makeDefHistory() (*GrueHistory, error) {
	var feeds = make(map[string]*RSSFeed)
	var hist = &GrueHistory{Feeds: feeds}
	return hist, nil
}

func writeDefHistory(path string) (*GrueHistory, error) {
	hist, err := makeDefHistory()
	if err != nil {
		return nil, err
	}
	hist.path = path
	return hist, hist.Write()
}

func getHistoryPath() string {
	dataPath := os.Getenv("XDG_DATA_HOME")
	if dataPath == "" {
		home := os.Getenv("HOME")
		if home == "" {
			panic("Can't find path to data directory")
		}
		return path.Join(os.Getenv("HOME"), ".local/share", "grue.json")
	}
	return path.Join(dataPath, "grue.json")
}

func ReadHistory() (*GrueHistory, error) {
	var hist *GrueHistory = new(GrueHistory)
	var path = getHistoryPath()
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return writeDefHistory(path)
	} else if err != nil {
		return nil, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(hist)
	if err != nil {
		return nil, err
	}
	hist.path = path
	return hist, nil
}
