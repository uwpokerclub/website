package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Pagination struct {
	Limit  *int `json:"limit,omitempty"  form:"limit,omitempty"  query:"limit,omitempty"`
	Offset *int `json:"offset,omitempty" form:"offset,omitempty" query:"offset,omitempty"`
}

const MaxLimit = 100

func ParsePagination(ctx *gin.Context) (Pagination, error) {
	var pagination Pagination
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		return Pagination{}, err
	}

	if pagination.Limit != nil {
		if *pagination.Limit <= 0 {
			defaultLimit := MaxLimit
			pagination.Limit = &defaultLimit
		} else if *pagination.Limit > MaxLimit {
			capped := MaxLimit
			pagination.Limit = &capped
		}
	}

	if pagination.Offset != nil && *pagination.Offset < 0 {
		zero := 0
		pagination.Offset = &zero
	}

	return pagination, nil
}

func (p *Pagination) Apply(query *gorm.DB) *gorm.DB {
	if p.Limit != nil {
		query = query.Limit(*p.Limit)
	}

	if p.Offset != nil {
		query = query.Offset(*p.Offset)
	}

	return query
}

