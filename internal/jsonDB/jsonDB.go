package jsonDB

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var ErrAlreadyExists = errors.New("already exists")
var ErrDoesNotExists = errors.New("does not exist")

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"user"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string
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

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	ds, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	highestID := 0
	for _, chirp := range ds.Chirps {
		if chirp.Id > highestID {
			highestID = chirp.Id
		}
	}

	chirp := Chirp{
		Id:   highestID + 1,
		Body: body,
	}

	ds.Chirps[chirp.Id] = chirp

	err = db.writeDB(ds)
	if err != nil {
		return chirp, fmt.Errorf("failed to write to database: %s", err)
	}

	return chirp, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateUser(email string, password string) (User, error) {
	ds, err := db.loadDB()
	if err != nil {
		return User{}, fmt.Errorf("failed to load database: %s", err)
	}

	highestID := 0
	for _, user := range ds.Users {
		if email == user.Email {
			return User{}, ErrAlreadyExists
		}
		if user.Id > highestID {
			highestID = user.Id
		}
	}

	user := User{
		Id:       highestID + 1,
		Email:    email,
		Password: password,
	}

	ds.Users[user.Id] = user

	err = db.writeDB(ds)
	if err != nil {
		return user, fmt.Errorf("failed to write to database: %s", err)
	}

	return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	ds, err := db.loadDB()
	if err != nil {
		return User{}, fmt.Errorf("failed to load database: %s", err)
	}

	for _, user := range ds.Users {
		if email == user.Email {
			return user, nil
		}
	}

	// user not found
	return User{}, ErrDoesNotExists

}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	file, err := os.Open(db.path)
	if err != nil {
		return DBStructure{}, fmt.Errorf("can't open db file: %s", err)
	}

	// Read the contents of the file into a byte slice
	stat, err := file.Stat()
	if err != nil {
		return DBStructure{}, fmt.Errorf("failed to turn data into byte slice %s", err)
	}
	bytes := make([]byte, stat.Size())

	_, err = file.Read(bytes)
	if err != nil {
		return DBStructure{}, fmt.Errorf("failed to read data: %s", err)
	}
	file.Close()

	ds := DBStructure{}
	if len(bytes) == 0 {
		ds.Chirps = map[int]Chirp{}
		ds.Users = map[int]User{}
		return ds, nil
	}

	// Unmarshal the JSON
	err = json.Unmarshal(bytes, &ds)
	if err != nil {
		return DBStructure{}, fmt.Errorf("failed to unmarshal JSON: %s", err)
	}

	return ds, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {

	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %s", err)
	}

	err = os.Truncate(db.path, 0)
	if err != nil {
		return fmt.Errorf("error truncating file: %s", err)
	}

	err = os.WriteFile(db.path, dat, 0644)
	if err != nil {
		return fmt.Errorf("error writing to database: %s", err)
	}

	return nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	ds, err := db.loadDB()
	if err != nil {
		return nil, fmt.Errorf("error loading the database: %s", err)
	}

	chirps := make([]Chirp, 0, len(ds.Chirps))
	for _, chirp := range ds.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

// GetChirp returns chirp with a specific ID
func (db *DB) GetChirp(id int) (Chirp, error) {
	ds, err := db.loadDB()
	if err != nil {
		return Chirp{}, fmt.Errorf("error loading the database: %s", err)
	}

	chirp, ok := ds.Chirps[id]
	if !ok {
		return Chirp{}, fmt.Errorf("chirp not found: %s", err)
	}

	return chirp, nil
}
