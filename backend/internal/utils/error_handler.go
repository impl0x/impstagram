package mo

// this file is mostly copied from the default error handler present in mo.
import (
	"backend/internal/pkg/errs"
	"backend/internal/pkg/responses"
	"backend/internal/utils/errorcodes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/impl0x/mo"
	"github.com/impl0x/mo/modules/logger"
	"github.com/impl0x/mo/validator"
)

type validationErrorJson struct {
	responses.Response
	Errors []validator.ValidationErrorJson `json:"errors,omitempty"`
}

// This function is given to the Mo framework which then bubbles all the errors here and we handle it here.
func CustomErrorHandler(c *mo.Context, err error) {
	if c.Response().IsCommitted() {
		if err == nil { // it means no error occurred and response was successfully written
			return
		}
		// if response is already committed then we cannot write a response anymore
		logger.Error("cannot write error, response already sent!", "err", err.Error())
		return
	}
	// if response was not written then we need to see if error exists and return appropriate messages
	switch e := err.(type) {
	case nil: // if no error was returned it means handlers probably returned a nil without writing a response
		c.NoContent(http.StatusNoContent)
	case mo.HTTPError: // the framework returned a http error, this happens only on routing errors and therefore only not_found and method_not_allowed
		// the error struct here is actually json compatible but we want to return errors in our schema so we will format it
		var codeName errorcodes.Code
		switch e.StatusCode() {
		case http.StatusNotFound:
			codeName = errorcodes.NotFound
		case http.StatusMethodNotAllowed:
			codeName = errorcodes.MethodNotAllowed
		default:
			codeName = errorcodes.Unknown // impossible logical case unless framework changes and somehow returns a different [mo.HTTPError]
		}
		c.JSON(e.StatusCode(), responses.Error(codeName, e.Error())) // e.Error returns the statusText of the statusCode if its a [HttpError]
	case *validator.GroupedValidationError:
		c.JSON(
			http.StatusBadRequest,
			validationErrorJson{
				responses.Error(errorcodes.ValidationError, e.Error()),
				e.ToJsonStructList(),
			},
		)
	case *json.SyntaxError:
		c.JSON(http.StatusUnprocessableEntity, responses.Error(
			errorcodes.InvalidJSON,
			fmt.Sprintf("JSON syntax error at offset %d", e.Offset),
		))
	case *json.UnmarshalTypeError:
		c.JSON(http.StatusExpectationFailed, responses.Error(
			errorcodes.ValidationError,
			fmt.Sprintf("Wrong type used for field %s", e.Field),
		))
	default:
		if e.Error() == "EOF" { // rare case because json parsing returns a errorString of EOF when a body is empty.
			c.JSON(http.StatusUnprocessableEntity, responses.Error(
				errorcodes.EOF,
				"End of file",
			))
			return
		}
		// we log internal errors
		logger.Error("Internal error: "+e.Error(), "errorType", fmt.Sprintf("%T", e))

		c.JSON(
			http.StatusInternalServerError,
			responses.Error(
				errorcodes.InternalServerError,
				errs.InternalServerError.Error(),
			),
		)
	}
}
