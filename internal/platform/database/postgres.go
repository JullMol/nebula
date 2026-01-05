package database

import (
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Job struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Image     string    `json:"image"`
	Command   string    `json:"command"`
	Status    string    `json:"status"` 
	Result    string    `json:"result"` 
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewConnection(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Job{})
	if err != nil {
		return nil, err
	}

	return db, nil
}