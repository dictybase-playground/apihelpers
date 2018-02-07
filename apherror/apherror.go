// Package apherror provides a consistent way to handler JSONAPI
// related http errors.
package apherror

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/errors/errhttp"
)

var (
	//HTTPError represents generic http errors
	Errhttp = errors.NewClass("HTTP Error", errors.NoCaptureStack())
	//ErrReqContext represents error in extracting context value from http request
	ErrReqContext = newErrorClass("Unable to retrieve context", http.StatusInternalServerError)
	//ErrOuthExchange represents error in exchanging code for token with oauth server
	ErrOauthExchange = newErrorClass("Unable to exchange token for code", http.StatusInternalServerError)
	//ErrUserRetrieval represents error in retrieving user information from oauth provider
	ErrUserRetrieval = newErrorClass("Unable to retrieve user information", http.StatusInternalServerError)
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
	//ErrSparseFieldSets represents any error with invalid sparse fieldsets query parameter
	ErrSparseFieldSets = newErrorClassWithParam("Invalid sparse fieldsets query parameter", "fieldsets", http.StatusBadRequest)
	//ErrFilterParam represents any error with invalid filter query paramter
	ErrFilterParam = newErrorClassWithParam("Invalid filter query parameter", "filter", http.StatusBadRequest)
	//ErrNotAcceptable represents any error with wrong or inappropriate http Accept header
	ErrNotAcceptable = newErrorClass("Accept header is not acceptable", http.StatusNotAcceptable)
	//ErrUnsupportedMedia represents any error with unsupported media type in http header
	ErrUnsupportedMedia = newErrorClass("Media type is not supported", http.StatusUnsupportedMediaType)
	//ErrQueryParam represents any error with http query parameters
	ErrQueryParam = newErrorClass("Invalid query parameter", http.StatusBadRequest)
	titleErrKey   = errors.GenSym()
	pointerErrKey = errors.GenSym()
	paramErrKey   = errors.GenSym()
)

// HTTPError is used for errors
type HTTPError struct {
	err    error
	msg    string
	status int
	Errors []Error `json:"errors,omitempty"`
}

// Error can be used for all kind of application errors
// e.g. you would use it to define form errors or any
// other semantical application problems
// for more information see http://jsonapi.org/format/#errors
type Error struct {
	ID     string       `json:"id,omitempty"`
	Links  *ErrorLinks  `json:"links,omitempty"`
	Status string       `json:"status,omitempty"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
	Meta   interface{}  `json:"meta,omitempty"`
}

// ErrorLinks is used to provide an About URL that leads to
// further details about the particular occurrence of the problem.
//
// for more information see http://jsonapi.org/format/#error-objects
type ErrorLinks struct {
	About string `json:"about,omitempty"`
}

// ErrorSource is used to provide references to the source of an error.
//
// The Pointer is a JSON Pointer to the associated entity in the request
// document.
// The Paramter is a string indicating which query parameter caused the error.
//
// for more information see http://jsonapi.org/format/#error-objects
type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

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
	err := Errhttp.NewClass(
		http.StatusText(code),
		errhttp.SetStatusCode(code),
	)
	err.MustAddData(titleErrKey, msg)
	return err
}

//JSONAPIError generate JSONAPI formatted http error from an error object
func JSONAPIError(w http.ResponseWriter, err error) {
	status := errhttp.GetStatusCode(err, http.StatusInternalServerError)
	title, _ := errors.GetData(err, titleErrKey).(string)
	jsnErr := Error{
		Status: strconv.Itoa(status),
		Title:  title,
		Detail: errhttp.GetErrorBody(err),
		Meta: map[string]interface{}{
			"creator": "api error helper",
		},
	}
	errSource := new(ErrorSource)
	pointer, ok := errors.GetData(err, pointerErrKey).(string)
	if ok {
		errSource.Pointer = pointer
	}
	param, ok := errors.GetData(err, paramErrKey).(string)
	if ok {
		errSource.Parameter = param
	}
	jsnErr.Source = errSource
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	encErr := json.NewEncoder(w).Encode(HTTPError{Errors: []Error{jsnErr}})
	if encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

// DatabaseError is for generating JSONAPI formatted error for database related
// errors
//func DatabaseError(w http.ResponseWriter, err error) {
//if err == dbr.ErrNotFound {
//JSONAPIError(w, ErrNotFound.New(err.Error()))
//return
//}
//// possible database query error
//JSONAPIError(w, ErrDatabaseQuery.New(err.Error()))
//}
