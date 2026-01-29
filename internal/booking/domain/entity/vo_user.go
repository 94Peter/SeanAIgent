package entity

import "errors"

var emptyUser = User{}

func NewUser(id, name string) (User, error) {
	if id == "" {
		return emptyUser, ErrUserInvalidID
	}
	if name == "" {
		return emptyUser, ErrUserInvalidName
	}
	return User{userID: id, userName: name}, nil
}

type User struct {
	userID   string
	userName string
}

func (u User) UserID() string {
	return u.userID
}

func (u User) UserName() string {
	return u.userName
}

var (
	ErrUserInvalidID   = errors.New("USER_INVALID_ID")
	ErrUserInvalidName = errors.New("USER_INVALID_NAME")
)
