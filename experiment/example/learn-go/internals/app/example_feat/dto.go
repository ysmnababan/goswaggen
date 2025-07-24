package example_feat

type UserCreateRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *UserResponse) MapFromUserModel(user *UserModel) {
	u.Name = user.Name
	u.Email = user.Email
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginResponse struct {
	TokenString string `json:"token"`
}
