package models

type Structure struct {
	ID     uint64  `json:"id"`
	Name   string  `json:"name"`
	Blinds []Blind `json:"blinds"`
}

type Blind struct {
	ID          uint64
	Small       int32
	Big         int32
	Ante        int32
	Time        int8
	Index       int
	StructureId uint64
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
	ID     uint64
	Name   string      `json:"name" binding:"required"`
	Blinds []BlindJSON `json:"blinds" binding:"required,dive"`
}
