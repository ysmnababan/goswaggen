package example_feat

import (
	"fmt"
	"learn-go/internals/factory"
	"learn-go/internals/utils/response"

	"learn-go/internals/utils/token"

	"errors"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type IUserRepo interface {
	Get(ctx echo.Context) (out []*UserModel, err error)
	Create(ctx echo.Context, in *UserModel) (out *UserModel, err error)
	GetByEmail(ctx echo.Context, email string) (out *UserModel, err error)
}

type userService struct {
	userRepo IUserRepo
}

func NewService(f *factory.Factory) *userService {
	return &userService{
		userRepo: NewRepo(f.Db),
	}
}

func (s *userService) Get(ctx echo.Context) (out []*UserResponse, err error) {
	out = []*UserResponse{}
	users, err := s.userRepo.Get(ctx)
	if err != nil {
		err = response.Wrap(response.ErrInternalServerError, fmt.Errorf("error repo: %w", err))
		return
	}

	for _, val := range users {
		user := &UserResponse{}
		user.MapFromUserModel(val)
		out = append(out, user)
	}
	return
}

func (s *userService) Create(ctx echo.Context, in *UserCreateRequest) (err error) {
	user := &UserModel{}
	user.Name = in.Name
	user.Email = in.Email
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPwd)
	_, err = s.userRepo.Create(ctx, user)
	if err != nil {
		err = response.Wrap(response.ErrInternalServerError, fmt.Errorf("error repo: %w", err))
		return
	}
	return nil
}

func (s *userService) Login(ctx echo.Context, req *UserLoginRequest) (out *UserLoginResponse, err error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		err = response.Wrap(response.ErrInternalServerError, fmt.Errorf("error repo: %w", err))
		return
	}

	if user.UserID == "" {
		err = response.Wrap(response.ErrInvalidUserCredentials, errors.New("invalid email or password"))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		err = response.Wrap(response.ErrInvalidUserCredentials, errors.New("invalid password or email"))
		return
	}

	// generate jwt token

	stringToken, err := token.GenerateJWT(user.UserID, user.Email)
	if err != nil {
		err = response.Wrap(response.ErrInternalServerError, fmt.Errorf("error creating token: %w", err))
		return
	}

	out = &UserLoginResponse{
		TokenString: stringToken,
	}
	return
}
