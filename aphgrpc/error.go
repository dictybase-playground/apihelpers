// Package aphgrpc provides various interfaces, functions, types
// for building and working with gRPC services.
package aphgrpc

import (
	"encoding/json"
	"net/http"
	"strconv"

	context "golang.org/x/net/context"
	dat "gopkg.in/mgutz/dat.v1"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/manyminds/api2go"
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
)

func newErrorWithParam(msg, param string) metadata.MD {
	return metadata.Pairs(MetaKey, msg, param)
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
	jsnErr := api2go.Error{
		Status: strconv.Itoa(status),
		Title:  md["error"][0],
		Detail: s.Message(),
		Meta: map[string]interface{}{
			"creator": "api error helper",
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	encErr := json.NewEncoder(w).Encode(api2go.HTTPError{Errors: []api2go.Error{jsnErr}})
	if encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

func fallbackError(w http.ResponseWriter, s *status.Status) {
	status := runtime.HTTPStatusFromCode(s.Code())
	jsnErr := api2go.Error{
		Status: strconv.Itoa(status),
		Title:  "gRPC error",
		Detail: s.Message(),
		Meta: map[string]interface{}{
			"creator": "api error helper",
		},
	}
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	encErr := json.NewEncoder(w).Encode(api2go.HTTPError{Errors: []api2go.Error{jsnErr}})
	if encErr != nil {
		http.Error(w, encErr.Error(), http.StatusInternalServerError)
	}
}

func handleError(ctx context.Context, err error) error {
	if err == dat.ErrNotFound {
		grpc.SetTrailer(ctx, ErrNotFound)
		return status.Error(codes.NotFound, err.Error())
	}
	grpc.SetTrailer(ctx, ErrDatabaseQuery)
	return status.Error(codes.Internal, err.Error())
}
