package models

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Plaintext string
	Hash      []byte
}

func (p *Password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 13)
	if err != nil {
		return err
	}

	p.Plaintext = plainTextPassword
	p.Hash = hash

	return nil
}

func (p *Password) Match(plainTextPassword string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plainTextPassword)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err // Internal server error
		}
	}
	return true, nil
}
