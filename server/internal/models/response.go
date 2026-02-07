package models

type ListResponse[T any] struct {
	Data  []T   `json:"data"`
	Total int64 `json:"total"`
}
