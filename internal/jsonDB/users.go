package jsonDB

import "fmt"

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

func (db *DB) UpdateUser(userId int, newEmail, newPassword string) (User, error) {
	ds, err := db.loadDB()
	if err != nil {
		return User{}, fmt.Errorf("failed to load database: %s", err)
	}

	user, ok := ds.Users[userId]
	if !ok {
		return User{}, ErrDoesNotExists
	}

	user.Email = newEmail
	user.Password = newPassword

	ds.Users[userId] = user

	err = db.writeDB(ds)
	if err != nil {
		return User{}, fmt.Errorf("failed to save user in database: %s", err)
	}

	return user, nil
}
