package main

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Server struct {
	ID        int64 `gorm:"type:bigserial;primary_key"`
	CreatedAt time.Time
	Hostname  string `gorm:"not null"`
}

func serverList(db *gorm.DB) ([]Server, error) {
	var servers []Server
	err := db.Order("hostname asc").Find(&servers).Error
	return servers, err
}

func serverGet(db *gorm.DB, hostname string) (*Server, error) {
	var server Server
	err := db.Where("hostname = ?", hostname).Find(&server).Error
	return &server, err
}
