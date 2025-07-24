package example_feat

import (
	"fmt"
	"learn-go/internals/factory"
	"learn-go/internals/utils/response"

	"github.com/labstack/echo/v4"
)

type IUserService interface {
	Get(ctx echo.Context) (out []*UserResponse, err error)
	Create(ctx echo.Context, in *UserCreateRequest) (err error)
	Login(ctx echo.Context, req *UserLoginRequest) (out *UserLoginResponse, err error)
}

type handler struct {
	service IUserService
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

// @Summary Get List of User
// @Description Get list of User
// @Tags user
// @Produce json
// @Success 200 {object} response.APIResponse{data=[]UserResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Param Authorization header string true "Bearer Token"
// @Router /api/v1/users [get]
func (h *handler) GetUsers(c echo.Context) error {
	fmt.Println(c.Get("user_id"))
	res, err := h.service.Get(c)
	if err != nil {
		return err
	}
	return response.WithStatusOKResponse(res, c)
}

// @Summary Create User
// @Description Create new User
// @Tags user
// @Accept json
// @Produce json
// @Param payload body UserCreateRequest true "Payload"
// @Success 200 {object} response.APIResponse{data=string}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users [post]
func (h *handler) CreateUser(c echo.Context) error {
	req := &UserCreateRequest{}
	err := c.Bind(req)
	if err != nil {
		return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
	}

	err = c.Validate(req)
	if err != nil {
		return response.Wrap(response.ErrValidation, fmt.Errorf("error validation: %w", err))
	}
	err = h.service.Create(c, req)
	if err != nil {
		return err
	}

	return response.WithStatusOKResponse("success", c)
}

// @Summary Login
// @Description User Login
// @Tags user
// @Accept json
// @Produce json
// @Param payload body UserLoginRequest true "Payload"
// @Success 200 {object} response.APIResponse{data=UserLoginResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/v1/users/auth [post]
func (h *handler) Login(c echo.Context) error {
	req := &UserLoginRequest{}
	err := c.Bind(req)
	if err != nil {
		return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
	}

	err = c.Validate(req)
	if err != nil {
		return response.Wrap(response.ErrValidation, fmt.Errorf("error validation: %w", err))
	}

	res, err := h.service.Login(c, req)
	if err != nil {
		return err
	}

	return response.WithStatusOKResponse(res, c)
}
