package aphgrpc

import (
	"fmt"
	"strings"

	"github.com/dictyBase/apihelpers/aphcollection"
	"github.com/dictyBase/go-genproto/dictybaseapis/api/jsonapi"
	"google.golang.org/grpc/metadata"
)

// JSONAPIParams is container for various JSON API query parameters
type JSONAPIParams struct {
	// contain include query paramters
	Includes []string
	// contain fields query paramters
	Fields []string
	// check for presence of fields parameters
	HasFields bool
	// check for presence of include parameters
	HasIncludes bool
}

func hasInclude(r *jsonapi.GetRequest) bool {
	if len(r.Include) > 0 {
		return true
	}
	return false
}

func hasFields(r *jsonapi.GetRequest) bool {
	if len(r.Fields) > 0 {
		return true
	}
	return false
}

// ValidateAndParseGetParams validate and parse the JSON API include and fields parameters
// that are used for singular resources
func ValidateAndParseGetParams(jsapi JSONAPIAllowedParams, r *jsonapi.GetRequest) (*JSONAPIParams, metadata.MD, error) {
	params := &JSONAPIParams{}
	if hasInclude(r) {
		if strings.Contains(r.Include, ",") {
			params.Includes = strings.Split(r.Include, ",")
		} else {
			params.Includes = []string{r.Include}
		}
		for _, v := range params.Includes {
			if !aphcollection.Contains(jsapi.AllowedInclude(), v) {
				return params, ErrIncludeParam, fmt.Errorf("include %s relationship is not allowed", v)
			}
		}
	} else {
		params.HasIncludes = false
	}

	if hasFields(r) {
		if strings.Contains(r.Fields, ",") {
			params.Fields = strings.Split(r.Fields, ",")
		} else {
			params.Fields = []string{r.Fields}
		}
		for _, v := range params.Fields {
			if !aphcollection.Contains(jsapi.AllowedFields(), v) {
				return params, ErrFilterParam, fmt.Errorf("%s value in fields is not allowed", v)
			}
		}
	} else {
		params.HasFields = true
	}
	return params, metadata.Pairs("errors", "none"), nil
}
