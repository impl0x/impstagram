package utils

// this file is mostly copied from the default error handler present in mo.
import (
	"backend/internal/pkg/apperr"
	"backend/internal/pkg/response"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/impl0x/mo"
	"github.com/impl0x/mo/modules/logger"
	"github.com/impl0x/mo/validator"
)

type validationErrorJson struct {
	response.Response
	Errors []validator.ValidationErrorJson `json:"errors,omitempty"`
}

func NewValidationErrorJson(response response.Response, err validator.GroupedValidationError) validationErrorJson {
	return validationErrorJson{response, err.ToJsonStructList()}
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
	// declaring some types we will need for errors.As
	var moHttpErr mo.HTTPError
	var validationErr validator.GroupedValidationError
	//// var jsonSyntaxErr *json.SyntaxError
	var jsonUnmarshalErr *json.UnmarshalTypeError
	var appErr apperr.AppErr

	switch {
	case errors.As(err, &appErr):
		c.JSON(appErr.Kind.ToStatusCode(), response.Error(appErr.Code, appErr.Message))
	case err == nil: // if no error was returned it means handlers probably returned a nil without writing a response
		c.NoContent(http.StatusNoContent)
	case errors.Is(err, context.Canceled):
	case errors.Is(err, context.DeadlineExceeded):
		c.JSON(http.StatusGatewayTimeout, response.Error(response.CodeTimeout, "Request timed out"))
	case errors.As(err, &moHttpErr): // the framework returned a http error, this happens only on routing errors and therefore only not_found and method_not_allowed
		// the error struct here is actually json compatible but we want to return errors in our schema so we will format it
		var codeName response.Code
		switch moHttpErr.StatusCode() {
		case http.StatusNotFound:
			codeName = response.CodeNotFound
		case http.StatusMethodNotAllowed:
			codeName = response.CodeMethodNotAllowed
		default:
			println("error handler: Unknown mo.HTTPError arrived, " + moHttpErr.Error())
			codeName = response.CodeUnknown // impossible logical case unless framework changes and somehow returns a different [mo.HTTPError] or we use mo.NewHTTPError in our own code. which we won't
		}
		c.JSON(moHttpErr.StatusCode(), response.Error(codeName, http.StatusText(moHttpErr.StatusCode()))) // e.Error returns the statusText of the statusCode if its a [HttpError]
	case errors.As(err, &validationErr): // returned by validator package in mo
		c.JSON(
			http.StatusBadRequest,
			validationErrorJson{
				response.Error(response.CodeValidationError, "Validation error, missing data or incorrect data"),
				validationErr.ToJsonStructList(),
			},
		)
	// 	it seems that json unmarshaler always returns a json.UnmarshalTypeError no matter if the syntax or type is incorrect, so this case will be commented out as being depreciated
	// // case errors.As(err, &jsonSyntaxErr):
	// // 	c.JSON(http.StatusUnprocessableEntity, response.Error(
	// // 		response.CodeJSONInvalid,
	// // 		fmt.Sprintf("JSON syntax error at offset %d", jsonSyntaxErr.Offset),
	// // 	))
	case errors.As(err, &jsonUnmarshalErr):
		c.JSON(http.StatusBadRequest, response.Error(
			response.CodeValidationError,
			"Invalid JSON",
		))
	case errors.Is(err, io.EOF):
		c.JSON(http.StatusUnprocessableEntity, response.Error(
			response.CodeEOF,
			"End of file",
		))
		return
	default:
		// we log internal errors
		logger.Error("Internal error: "+err.Error(), "errorType", fmt.Sprintf("%T", err))
		c.JSON(
			http.StatusInternalServerError,
			response.Error(
				response.CodeInternal,
				"Internal server error",
			),
		)
	}
}
