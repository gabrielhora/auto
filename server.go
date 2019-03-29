package main

import (
	"github.com/jinzhu/gorm"
	"os"
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

func serverGet(db *gorm.DB, hostname string) (Server, error) {
	var server Server
	err := db.Where("hostname = ?", hostname).Find(&server).Error
	if gorm.IsRecordNotFoundError(err) {
		return Server{}, nil
	}
	return server, err
}

func serverRegisterSelf(db *gorm.DB) (Server, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return Server{}, err
	}

	s, err := serverGet(db, hostname)
	if err != nil {
		return Server{}, err
	}
	if s.ID > 0 {
		return s, nil
	}

	s = Server{Hostname: hostname}
	err = db.Create(&s).Error
	return s, err
}
