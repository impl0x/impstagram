package auth

import "github.com/google/uuid"

type User struct { // not complete, will do after i setup database properly
	Id uuid.UUID
	Username string
	Email string
	Phone string
	PasswordHash string
	Dob string

}

func NewUser(req RegisterRequest, passwordHash string)User{
	return User{
		Username: req.Username,
		Email: req.Email,
		Phone: req.Phone,
		PasswordHash: passwordHash,
		Dob: req.Dob,
	}
}