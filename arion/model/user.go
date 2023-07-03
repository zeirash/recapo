package model

import (
	"github.com/gofrs/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"`
	Name     string    `db:"name"`
	Email    string `db:"email"`
	Password string `db:"password"`
}
