package models

type Structure struct {
	ID     uint    `json:"id" gorm:"type:serial;primaryKey"`
	Name   string  `json:"name" gorm:"not null"`
	Blinds []Blind `json:"blinds"`
}

type Blind struct {
	ID          uint  `json:"-" gorm:"type:serial;primaryKey"`
	Small       int32 `json:"small" gorm:"not null"`
	Big         int32 `json:"big" gorm:"not null"`
	Ante        int32 `json:"ante" gorm:"not null"`
	Time        int8  `json:"time" gorm:"not null"`
	Index       int8  `json:"-" gorm:"not null"`
	StructureId uint  `json:"-"`
}

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
	ID     uint
	Name   string      `json:"name" binding:"required"`
	Blinds []BlindJSON `json:"blinds" binding:"required,dive"`
}
