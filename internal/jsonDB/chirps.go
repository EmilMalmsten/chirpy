package jsonDB

import "fmt"

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, author_id int) (Chirp, error) {
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
		Id:       highestID + 1,
		Body:     body,
		AuthorId: author_id,
	}

	ds.Chirps[chirp.Id] = chirp

	err = db.writeDB(ds)
	if err != nil {
		return chirp, fmt.Errorf("failed to write to database: %s", err)
	}

	return chirp, nil
}

func (db *DB) DeleteChirp(chirp_id, user_id int) error {
	ds, err := db.loadDB()
	if err != nil {
		return err
	}

	_, ok := ds.Chirps[chirp_id]
	if !ok {
		return ErrDoesNotExists
	}

	if ds.Chirps[chirp_id].AuthorId != user_id {
		return ErrNotAuthorized
	}

	delete(ds.Chirps, chirp_id)

	err = db.writeDB(ds)
	if err != nil {
		return fmt.Errorf("failed to write to database: %s", err)
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
