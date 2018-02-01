// Package aphgrpc provides various interfaces, functions, types
// for building and working with gRPC services.
package aphgrpc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	context "golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// MetaKey is the key used for storing all metadata
	MetaKey = "error"
)

var (
	//ErrDatabaseQuery represents database query related errors
	ErrDatabaseQuery = newError("Database query error")
	//ErrDatabaseInsert represents database insert related errors
	ErrDatabaseInsert = newError("Database insert error")
	//ErrDatabaseUpdate represents database update related errors
	ErrDatabaseUpdate = newError("Database update error")
	//ErrDatabaseDelete represents database update delete errors
	ErrDatabaseDelete = newError("Database delete error")
	//ErrNotFound represents the absence of an HTTP resource
	ErrNotFound = newError("Resource not found")
	//ErrJSONEncoding represents any json encoding error
	ErrJSONEncoding = newError("Json encoding error")
	//ErrStructMarshal represents any error with marshalling structure
	ErrStructMarshal = newError("Structure marshalling error")
	//ErrIncludeParam represents any error with invalid include query parameter
	ErrIncludeParam = newErrorWithParam("Invalid include query parameter", "include")
	//ErrSparseFieldSets represents any error with invalid sparse fieldsets query parameter
	ErrFields = newErrorWithParam("Invalid field query parameter", "field")
	//ErrFilterParam represents any error with invalid filter query paramter
	ErrFilterParam = newErrorWithParam("Invalid filter query parameter", "filter")
	//ErrNotAcceptable represents any error with wrong or inappropriate http Accept header
	ErrNotAcceptable = newError("Accept header is not acceptable")
	//ErrUnsupportedMedia represents any error with unsupported media type in http header
	ErrUnsupportedMedia = newError("Media type is not supported")
	//ErrRetrieveMetadata represents any error to retrieve grpc metadata from the running context
	ErrRetrieveMetadata = errors.New("unable to retrieve metadata")
	//ErrXForwardedHost represents any failure or absence of x-forwarded-host HTTP header in the grpc context
	ErrXForwardedHost = errors.New("x-forwarded-host header is absent")
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

func newErrorWithParam(msg, param string) metadata.MD {
	return metadata.Pairs(MetaKey, msg, MetaKey, param)
}

func newError(msg string) metadata.MD {
	return metadata.Pairs(MetaKey, msg)
}

// CustomHTTPError is a custom error handler for grpc-gateway to generate
// JSONAPI formatted HTTP response.
func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		fallbackError(w, getgRPCStatus(errors.Wrap(err, "unable to retrieve metadata")))
		return
	}
	JSONAPIError(w, md.TrailerMD, getgRPCStatus(err))
}

func getgRPCStatus(err error) *status.Status {
	s, ok := status.FromError(err)
	if !ok {
		return status.New(codes.Unknown, err.Error())
	}
	return s
}

// JSONAPIError generates JSONAPI formatted error message
func JSONAPIError(w http.ResponseWriter, md metadata.MD, s *status.Status) {
	status := runtime.HTTPStatusFromCode(s.Code())
	jsnErr := Error{
		Status: strconv.Itoa(status),
		Title:  strings.Join(md["error"], "-"),
		Detail: s.Message(),
		Meta: map[string]interface{}{
			"creator": "api error helper",
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	encErr := json.NewEncoder(w).Encode(HTTPError{Errors: []Error{jsnErr}})
	if encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

func fallbackError(w http.ResponseWriter, s *status.Status) {
	status := runtime.HTTPStatusFromCode(s.Code())
	jsnErr := Error{
		Status: strconv.Itoa(status),
		Title:  "gRPC error",
		Detail: s.Message(),
		Meta: map[string]interface{}{
			"creator": "api error helper",
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	encErr := json.NewEncoder(w).Encode(HTTPError{Errors: []Error{jsnErr}})
	if encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

func HandleError(ctx context.Context, err error) error {
	switch err {
	case strings.Contains(err.Error(), "no rows"):
		grpc.SetTrailer(ctx, ErrNotFound)
		return status.Error(codes.NotFound, err.Error())
	case ErrRetrieveMetadata:
		grpc.SetTrailer(ctx, newError(err.Error()))
		return status.Error(codes.Internal, err.Error())
	case ErrXForwardedHost:
		grpc.SetTrailer(ctx, newError(err.Error()))
		return status.Error(codes.Internal, err.Error())
	default:
		grpc.SetTrailer(ctx, ErrDatabaseQuery)
		return status.Error(codes.Internal, err.Error())
	}
}
