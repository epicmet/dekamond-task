package users

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	PhoneNumber  string        `json:"phone_number" bson:"phone_number"`
	RegisteredAt time.Time     `json:"registered_at" bson:"registered_at"`
}

type PaginatedUsers struct {
	Users      []User `json:"users"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalCount int64  `json:"total_count"`
	TotalPages int    `json:"total_pages"`
}

type UserRepository interface {
	Create(phoneNumber string) (*User, error)
	FindByID(id string) (*User, error)
	FindByPhone(phoneNumber string) (*User, error)
	Upsert(phoneNumber string) (*User, error)
	Search(query string, page, pageSize int) (*PaginatedUsers, error)
	GetAll(page, pageSize int) (*PaginatedUsers, error)
}

var ErrUserNotFound = errors.New("user not found")
