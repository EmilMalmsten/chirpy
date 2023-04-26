package jsonDB

import (
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	id   int
	body string
}

func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	_, err := os.Create(db.path)
	if err != nil {
		return &db, err
	}
	return &db, nil
}
