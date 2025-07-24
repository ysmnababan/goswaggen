package example_feat

import (
	"time"

	"learn-go/internals/abstraction"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserBase struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-" gorm:"column:password"`
}

type UserModel struct {
	UserID string `json:"user_id" gorm:"primary_key;auto_increment"`
	UserBase
	abstraction.Entity
}

func (UserModel) TableName() string {
	return "users"
}

func (m *UserModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.CreatedAt = time.Now().UnixMilli()
	if m.UserID == "" {
		m.UserID = uuid.NewString()
	}

	return
}

func (m *UserModel) BeforeUpdate(tx *gorm.DB) (err error) {
	now := time.Now().UnixMilli()
	m.ModifiedAt = &now
	return
}
