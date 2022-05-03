package utils

import (
	uuid "github.com/satori/go.uuid"
)

// CreateUUID 创建一局的uuid
func CreateUUID() string {
	return uuid.NewV4().String()
}
