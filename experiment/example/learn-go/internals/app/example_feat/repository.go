package example_feat

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type userRepo struct {
	*gorm.DB
}

func NewRepo(db *gorm.DB) *userRepo {
	return &userRepo{DB: db}
}

func (r *userRepo) Get(ctx echo.Context) (out []*UserModel, err error) {
	out = []*UserModel{}
	res := r.DB.Find(&out)
	if res.Error != nil {
		err = res.Error
		return
	}
	return
}

func (r *userRepo) Create(ctx echo.Context, in *UserModel) (out *UserModel, err error) {
	res := r.DB.Create(in)
	if res.Error != nil {
		err = res.Error
		return
	}
	out = in
	return
}

func (r *userRepo) GetByEmail(ctx echo.Context, email string) (out *UserModel, err error) {
	out = &UserModel{}
	res := r.DB.Where("email = ?", email).First(&out)
	if res.Error != nil && res.Error != gorm.ErrRecordNotFound {
		err = res.Error
		return
	}
	return
}
