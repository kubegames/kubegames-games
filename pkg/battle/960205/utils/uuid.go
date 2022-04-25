package utils

import (
	"common/log"

	uuid "github.com/satori/go.uuid"
)

// CreateUUID 创建一局的uuid
func CreateUUID() string {
	u1, err := uuid.NewV4()
	if err != nil {
		log.Errorf("牌局uuid生成出错:%s", err.Error())
	}
	return u1.String()
}
