// Package apherror provides a consistent way to handler JSONAPI
// related http errors.
package apherror

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gocraft/dbr"

	"github.com/manyminds/api2go"
	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
)

var (
	//HTTPError represents generic http errors
	HTTPError = errors.NewClass("HTTP Error", errors.NoCaptureStack())
	//ErrDatabaseQuery represents database query related errors
	ErrDatabaseQuery = newErrorClass("Database query error", http.StatusInternalServerError)
	//ErrNotFound represents the absence of an HTTP resource
	ErrNotFound = newErrorClass("Resource not found", http.StatusNotFound)
	//ErrJSONEncoding represents any json encoding error
	ErrJSONEncoding = newErrorClass("Json encoding error", http.StatusInternalServerError)
	//ErrStructMarshal represents any error with marshalling structure
	ErrStructMarshal = newErrorClass("Structure marshalling error", http.StatusInternalServerError)
	//ErrIncludeParam represents any error with invalid include query parameter
	ErrIncludeParam = newErrorClassWithParam("Invalid include query parameter", "include", http.StatusBadRequest)
	//ErrNotAcceptable represents any error with wrong or inappropriate http Accept header
	ErrNotAcceptable = newErrorClass("Accept header is not acceptable", http.StatusNotAcceptable)
	//ErrUnsupportedMedia represents any error with unsupported media type in http header
	ErrUnsupportedMedia = newErrorClass("Media type is not supported", http.StatusUnsupportedMediaType)
	titleErrKey         = errors.GenSym()
	pointerErrKey       = errors.GenSym()
	paramErrKey         = errors.GenSym()
)

func newErrorClassWithParam(msg, param string, code int) *errors.ErrorClass {
	err := newErrorClass(msg, code)
	err.MustAddData(paramErrKey, param)
	return err
}

func newErrorClassWithPointer(msg, pointer string, code int) *errors.ErrorClass {
	err := newErrorClass(msg, code)
	err.MustAddData(pointerErrKey, pointer)
	return err
}

func newErrorClass(msg string, code int) *errors.ErrorClass {
	err := HTTPError.NewClass(
		http.StatusText(code),
		errhttp.SetStatusCode(code),
	)
	err.MustAddData(titleErrKey, msg)
	return err
}

//JSONAPIError generate JSONAPI formatted http error from an error object
func JSONAPIError(w http.ResponseWriter, err error) {
	title, _ := errors.GetData(err, titleErrKey).(string)
	jsnErr := api2go.Error{
		Status: strconv.Itoa(errhttp.GetStatusCode(err, http.StatusInternalServerError)),
		Title:  title,
		Detail: errhttp.GetErrorBody(err),
		Meta: map[string]interface{}{
			"creator": "modware api",
		},
	}
	pointer, ok := errors.GetData(err, pointerErrKey).(string)
	if ok {
		jsnErr.Source.Pointer = pointer
	}
	param, ok := errors.GetData(err, paramErrKey).(string)
	if ok {
		jsnErr.Source.Parameter = param
	}
	encErr := json.NewEncoder(w).Encode(api2go.HTTPError{Errors: []api2go.Error{jsnErr}})
	if encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

// DatabaseError is for generating JSONAPI formatted error for database related
// errors
func DatabaseError(w http.ResponseWriter, err error) {
	if err == dbr.ErrNotFound {
		JSONAPIError(w, ErrNotFound.New(err.Error()))
		return
	}
	// possible database query error
	JSONAPIError(w, ErrDatabaseQuery.New(err.Error()))
}
