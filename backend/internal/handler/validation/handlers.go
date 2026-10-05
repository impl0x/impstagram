package validations

import (
	"backend/internal/validation"
	"backend/pkg/response"
	"net/http"

	"github.com/impl0x/mo"
)

// Registers all the paths for validation group,
// it is recommended to use this function instead of registering paths one by one yourself.
//
// paths:
//   - POST - /username
//   - GET - /username/rules
//   - POST - /dob
//   - GET - /dob/rules
func RegisterPaths(g mo.Grouped) {
	g.POST("/username", Username)
	g.GET("/username/rules", UsernameRules)
}

// wrapper for validation endpoint handlers
func validateStrHandler(fn func(string) error) mo.HandlerFunc {
	return func(c *mo.Context) error {
		var v strValue
		err := c.DecodeAndValidateBody(&v)
		if err != nil {
			return err
		}
		err = fn(v.Value)
		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				response.Error(
					response.CodeValidationError,
					err.Error(),
				),
			)
		}
		return c.NoContent(http.StatusNoContent)
	}
}

// wrapper for /rules endpoint handlers
func ruleHandler(name string, rules []string) mo.HandlerFunc {
	return func(c *mo.Context) error {
		return c.JSON(
			http.StatusOK,
			response.Success(
				response.CodeOk,
				"Following are the rules for "+name+" validation",
				struct {
					Rules []string `json:"rules"`
				}{rules},
			),
		)
	}
}

// Validates a username
//   - POST - [strValue]
func Username(c *mo.Context) error {
	return validateStrHandler(validation.Username)(c)
}

// Returns the rules for username validation
//   - GET
func UsernameRules(c *mo.Context) error {
	return ruleHandler("username", validation.UsernameRules)(c)
}

// Validates a dob
//   - POST - [strValue]
func Dob(c *mo.Context) error {
	return validateStrHandler(validation.Dob)(c)
}

// Returns the rules for dob validation
//   - GET
func DobRules(c *mo.Context) error {
	return ruleHandler("dob", validation.DobRules)(c)
}
