package mysql

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lorenzoranucci/eda-semi-async-http-handler/internal/application"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

type User struct {
	RequestUUID uuid.UUID `db:"request_uuid"`
	UserUUID    uuid.UUID `db:"uuid"`
	UserName    string    `db:"name"`
}

func (u *UserRepository) FindUserByName(name string) (*application.User, error) {
	var users []User
	err := u.db.Select(&users, "SELECT request_uuid, uuid, name FROM users WHERE name = ?", name)
	if err != nil {
		return nil, err
	}

	if users == nil || len(users) == 0 {
		return nil, nil
	}

	return &application.User{
		RequestUUID: users[0].RequestUUID,
		UserUUID:    users[0].UserUUID,
		UserName:    users[0].UserName,
	}, nil
}

func (u *UserRepository) AddUser(user *application.User) error {
	_, err := u.db.Exec("INSERT INTO users (request_uuid, uuid, name) VALUES (?, ?, ?)", user.RequestUUID.String(), user.UserUUID.String(), user.UserName)
	return err
}
