package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// User represented by the parsed JWT
type User struct {
	authenticated bool
	claims        jwt.Claims
	attributes    UserAttributes
}

// UserAttributes for User struct
type UserAttributes struct {
	Enabled          bool                 `json:"enabled"`
	Attributes       []UserAttributeField `json:"attributes"`
	CreatedDate      time.Time            `json:"createdDate"`
	LastModifiedDate time.Time            `json:"lastModifiedDate"`
	Username         string               `json:"username"`
	Status           string               `json:"status"`
}

// UserAttributeField name value keys
type UserAttributeField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
