package models

import "gorm.io/gorm"

type Structure struct {
	ID     int32   `json:"id" gorm:"type:integer;primaryKey;autoIncrement"`
	Name   string  `json:"name" gorm:"not null"`
	Blinds []Blind `json:"blinds" gorm:"foreignKey:StructureId"`
} //@name Structure

func (Structure) TableName() string {
	return "structures"
}

func (Structure) Preload(tx *gorm.DB) *gorm.DB {
	return tx.Preload("Blinds")
}

func (s *Structure) AfterSave(tx *gorm.DB) (err error) {
	err = s.Preload(tx).Find(s).Error
	return
}

type Blind struct {
	ID          int32 `json:"-" gorm:"type:integer;primaryKey;autoIncrement"`
	Small       int32 `json:"small" gorm:"not null"`
	Big         int32 `json:"big" gorm:"not null"`
	Ante        int32 `json:"ante" gorm:"not null"`
	Time        int8  `json:"time" gorm:"not null"`
	Index       int8  `json:"-" gorm:"not null"`
	StructureId int32 `json:"-" gorm:"type:integer;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}//@name Blind

type BlindJSON struct {
	Small int32 `json:"small" binding:"required,gte=0"`
	Big   int32 `json:"big" binding:"required,gte=0"`
	Ante  int32 `json:"ante" binding:"omitempty,gte=0"`
	Time  int8  `json:"time" binding:"required,gt=0,lte=60"`
}

type CreateStructureRequest struct {
	Name   string      `json:"name" binding:"required"`
	Blinds []BlindJSON `json:"blinds" binding:"required,dive"`
}

type UpdateStructureRequest struct {
	ID     int32
	Name   string      `json:"name" binding:"required"`
	Blinds []BlindJSON `json:"blinds" binding:"required,dive"`
}
