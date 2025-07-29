package example_feat

import "encoding/json"

type UserCreateRequest struct {
	Name        string            `json:"name" validate:"required"`     // basic
	Email       string            `json:"email" validate:"required"`    // basic
	Password    string            `json:"password" validate:"required"` // basic
	Birthdate   *string           `json:"birthdate"`                    // pointer to basic
	Personal    Personal          // named struct
	Metadata    map[string]string // map
	Tags        []string          // slice
	Scores      []int             // slice of int
	Reference   *OtherReference   // pointer to named type
	Misc        interface{}       // interface
	Raw         json.RawMessage   // alias for []byte
	Coordinates [2]float64        // array
	Callback    func(int) error   // function
	IsActive    bool              // basic
}

type OtherReference struct {
	ID   int
	Code string
}
type Personal struct {
	Age   int
	Hobby *string
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
