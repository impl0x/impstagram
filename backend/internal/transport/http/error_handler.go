package httpserver

// this file is mostly copied from the default error handler present in mo.
import (
	"backend/internal/transport/http/response"
	"backend/pkg/apperr"
	"context"
	"encoding/json"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
	"github.com/impl0x/mo/validator/v3"
)

type validationErrorJson struct {
	response.Response
	Errors []validator.ValidationErrorJson `json:"errors,omitempty"`
}

func NewValidationErrorJson(response response.Response, err validator.GroupedValidationError) validationErrorJson {
	return validationErrorJson{response, err.ToJsonStructList()}
}

// This function is given to the Mo framework which then bubbles all the errors here and we handle it here.
func ErrorHandler(log logger.Logger) mo.HTTPErrorHandler {
	return func(c *mo.Context, err error) {
		if c.Response().IsCommitted() {
			if err == nil { // it means no error occurred and response was successfully written
				return
			}
			// if response is already committed then we cannot write a response anymore
			log.Error("cannot write error, response already sent!", "err", err.Error())
			return
		}
		// if response was not written then we need to see if error exists and return appropriate messages
		// declaring some types we will need for errors.As
		var moHttpErr mo.HttpError
		var validationErr validator.GroupedValidationError
		var jsonSyntaxErr *jsontext.SyntacticError
		var jsonUnmarshalErr *json.UnmarshalTypeError
		var appErr apperr.AppErr

		switch {
		case err == nil: // if no error was returned it means handlers probably returned a nil without writing a response
			c.NoContent(http.StatusNoContent)
		case errors.As(err, &appErr):
			c.JSON(response.HttpStatuscodeFromAppErrKind(appErr.Kind), response.FromAppErr(appErr, nil))
		case errors.Is(err, context.Canceled):
		case errors.Is(err, context.DeadlineExceeded):
			c.JSON(http.StatusGatewayTimeout, response.Error(apperr.CodeTimeout, "Request timed out"))
		case errors.As(err, &moHttpErr): // the framework returned a http error, this happens only on routing errors and therefore only not_found and method_not_allowed
			// the error struct here is actually json compatible but we want to return errors in our schema so we will format it
			var codeName apperr.Code
			var message string
			switch moHttpErr.StatusCode() {
			case http.StatusNotFound:
				codeName = apperr.CodeNotFound
				message = "Path not found"
			case http.StatusMethodNotAllowed:
				codeName = apperr.CodeMethodNotAllowed
				message = "Method not allowed"
			default:
				log.Warn("Unknown mo.HttpError arrived, " + moHttpErr.Error())
				message = "An unknown and unexpected error has occurred"
				codeName = apperr.CodeUnknown // impossible logical case unless framework changes and somehow returns a different [mo.HTTPError] or we use mo.NewHTTPError in our own code. which we won't
			}
			c.JSON(moHttpErr.StatusCode(), response.Error(codeName, message)) // e.Error returns the statusText of the statusCode if its a [HttpError]
		case errors.As(err, &validationErr): // returned by validator package in mo
			c.JSON(
				http.StatusBadRequest,
				validationErrorJson{
					response.Error(apperr.CodeValidationError, "Validation error, missing data or incorrect data"),
					validationErr.ToJsonStructList(),
				},
			)
		case errors.As(err, &jsonSyntaxErr):
			c.JSON(http.StatusUnprocessableEntity, response.Error(
				apperr.CodeJSONInvalid,
				fmt.Sprintf("JSON syntax error at offset %d", jsonSyntaxErr.ByteOffset),
			))
		case errors.As(err, &jsonUnmarshalErr):
			c.JSON(http.StatusBadRequest, response.Error(
				apperr.CodeValidationError,
				"Invalid JSON",
			))
		case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
			c.JSON(http.StatusUnprocessableEntity, response.Error(
				apperr.CodeEOF,
				"End of file",
			))
			return
		default:
			// we log internal errors
			var args []string
			args = append(args, "error type", reflect.TypeOf(err).String())
			// get request id if present
			if reqId, e := c.GetTyped[string](KeyRequestID); e != nil {
				args = append(args, "req", reqId)
			}
			log.Error("Internal: "+err.Error(), args...)
			c.JSON(
				http.StatusInternalServerError,
				response.Error(
					apperr.CodeInternal,
					"Internal server error",
				),
			)
		}
	}
}
