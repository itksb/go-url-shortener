package user

import "github.com/google/uuid"

// This package is part of infrastructure layer, not domain.
// Put authentication related staff here

type userID string

// FieldID field name for user ID
const FieldID userID = "uid"

// GenerateUserID - generates unique user id using uuid
func GenerateUserID() string {
	return uuid.NewString()
}
