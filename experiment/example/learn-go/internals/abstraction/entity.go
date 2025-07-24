package abstraction

import (
	"time"

	"gorm.io/gorm"
)

type Entity struct {
	CreatedAt  int64  `json:"created_at"`
	CreatedBy  string `json:"created_by"`
	ModifiedAt *int64 `json:"modified_at"`
	ModifiedBy string `json:"modified_by"`

	DeletedAt *int64 `json:"deleted_at"`
	DeletedBy string `json:"deleted_by"`
}

type Filter struct {
	CreatedAt  *int64 `query:"created_at"`
	CreatedBy  *int   `query:"created_by"`
	ModifiedAt *int64 `query:"modified_at"`
	ModifiedBy *int   `query:"modified_by"`
}

func (m *Entity) BeforeUpdate(tx *gorm.DB) (err error) {
	ma := time.Now().UnixMilli()
	m.ModifiedAt = &ma
	return
}

func (m *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreatedAt = time.Now().UnixMilli()
	return
}
