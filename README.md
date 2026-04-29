# 🛠️ goswaggen

**goswaggen** is a command-line tool that helps you automatically generate Swagger (OpenAPI) comment annotations for your Go HTTP handlers.

It reduces the manual effort of writing documentation by analyzing your code and generating structured comments that tools like [`swaggo/swag`](https://github.com/swaggo/swag) can parse to build your API docs.

## 💡 Key Features

- Parses your Go handler functions
- Generates Swagger-compatible comment blocks above each handler
- Compatible with `Echo`.
- Simple CLI usage — integrates into your workflow

## 🚀 Getting Started

### Installation

#### For Go Developers

1. **Install the CLI:**

   ```bash
   go install github.com/ysmnababan/goswaggen@latest
   ```

2. **(Optional) Initialize config:**

   ```bash
   goswaggen init
   ```

   This creates a `goswaggen.yml` for customizing default error and success responses:

   ```yml
   error_response: "default.Error"
   success_response: "default.Success"
   security: "BearerAuth"
   ```

3. **Navigate to your project directory:**  
   Ensure your handlers are registered to a router in `main.go`.

4. **Generate Swagger comments:**

   ```bash
   # Generate for a specific handler
   goswaggen gen <handler_name>

   # Force direct insertion above handlers
   goswaggen gen <handler_name> -f
   ```

5. **Review and adjust generated comments as needed.**

#### For Non-Go Users

- Download pre-built binaries from the [GitHub Releases](https://github.com/ysmnababan/goswaggen/releases) page.

---

## 📝 Commands Overview

- `goswaggen list`  
  List all registered handlers and their packages.

- `goswaggen init`  
  Generate a `goswaggen.yml` config file.

- `goswaggen gen <handler_name>`  
  Output Swagger comment block for a handler to the CLI.  
  Handler name can be plain or qualified (e.g., `user.Login`).

- `goswaggen gen <handler_name> -f`  
  Directly insert comment block above the handler in your code.

**Examples:**

```bash
goswaggen gen Login
goswaggen gen user.Login
goswaggen gen Login -f
```

---

## 📚 Usage Examples

**Before:**

```go
func (h *handler) LoginAsCompanyAdmin(c echo.Context) error {
 ctx, ok := c.(*abstraction.Context)
 if !ok {
  return response.ErrorWrap(response.ErrBadRequest, errors.New("context invalid")).Send(c)
 }
 req := &dto.LoginAsCompanyAdminRequest{}
 if err := c.Bind(req); err != nil {
  return response.ErrorWrap(response.ErrValidation, err).Send(c)
 }
 if err := c.Validate(req); err != nil {
  return response.ErrorWrap(response.ErrUnprocessableEntity, err).Send(c)
 }
 data, err := h.authService.LoginAsCompanyAdmin(ctx, req)
 if err != nil {
  return response.ErrorResponse(err).Send(c)
 }
 return response.SuccessResponse(data).Send(c)
}
```

**After:**

By running the command `goswaggen gen LoginAsCompanyAdmin` :

```go
// @Summary LoginAsCompanyAdmin
// @Description Login as company admin
// @Tags ______
// @Accept json
// @Produce json
// @Param req body dto.LoginAsCompanyAdminRequest true "change this description"
// @Success 200 {object} response.Success{} "success"
// @Failure 500 {object} response.Error "error"
// @Failure 400 {object} response.Error "error"
// @Failure 404 {object} response.Error "error"
// @Router /app/v1/company/auth/login [post]
func (h *handler) LoginAsCompanyAdmin(c echo.Context) error {
 ctx, ok := c.(*abstraction.Context)
 if !ok {
  return response.ErrorWrap(response.ErrBadRequest, errors.New("context invalid")).Send(c)
 }
 req := &dto.LoginAsCompanyAdminRequest{}
 if err := c.Bind(req); err != nil {
  return response.ErrorWrap(response.ErrValidation, err).Send(c)
 }
 if err := c.Validate(req); err != nil {
  return response.ErrorWrap(response.ErrUnprocessableEntity, err).Send(c)
 }
 data, err := h.authService.LoginAsCompanyAdmin(ctx, req)
 if err != nil {
  return response.ErrorResponse(err).Send(c)
 }
 return response.SuccessResponse(data).Send(c)
}
```

## ❓ FAQ and Troubleshooting

See [FAQ](./docs/faq.md) for common questions.

---

_Tip: Keep your handlers registered and code organized for best results!_

