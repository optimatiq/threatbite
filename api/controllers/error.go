package controllers

import (
	"errors"
)

// ErrInvalidIP indicates that IP address is invalid
var ErrInvalidIP = errors.New("invalid IP address")

// ErrInvalidEmail indicates that email is invalid
var ErrInvalidEmail = errors.New("invalid email")
