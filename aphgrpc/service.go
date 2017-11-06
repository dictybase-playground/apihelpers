package aphgrpc

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/fatih/structs"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	context "golang.org/x/net/context"
)

const (
	DefaultPagenum  = 1
	DefaultPagesize = 10
)

// JSONAPIParamsInfo interface should be implement by all grpc-gateway services
// that supports JSON API specifications.
type JSONAPIParamsInfo interface {
	// Relationships that could be included
	AllowedInclude() []string
	// Attribute fields that are allowed
	AllowedFields() []string
	// Filter fields that are allowed
	AllowedFilter() []string
	// FilterToColumns provides mapping between filter and storage columns
	FilterToColumns() map[string]string
	// RequiredAttrs are the mandatory attributes for creating a new resource
	RequiredAttrs() []string
}

// JSONAPIResource interface provides information about HTTP resource. All
// grpc-gateway services that supports JSONAPI should implement this interface.
type JSONAPIResource interface {
	//GetResourceName returns canonical resource name
	GetResourceName() string
	// GetBaseURL returns the base url with the scheme
	GetBaseURL() string
	// GetPrefix returns the path that could be appended to base url
	GetPathPrefix() string
}

// GetTotalPageNum calculate total no of pages from total no. records and page size
func GetTotalPageNum(record, pagesize int64) int64 {
	total := int64(math.Floor(float64(record) / float64(pagesize)))
	if math.Mod(float64(record), float64(pagesize)) > 0 {
		total += 1
	}
	return total
}

// GetPaginatedLinks gets paginated links and total page number for collection resources
func GetPaginatedLinks(rs JSONAPIResource, lastpage, pagenum, pagesize int64) map[string]string {
	var links map[string]string
	links["self"] = GenPaginatedResourceLink(rs, pagenum, pagesize)
	links["first"] = GenPaginatedResourceLink(rs, 1, pagesize)
	if pagenum != 1 {
		links["previous"] = GenPaginatedResourceLink(rs, pagenum-1, pagesize)
	}
	links["last"] = GenPaginatedResourceLink(rs, lastpage, pagesize)
	if pagenum != lastpage {
		links["next"] = GenPaginatedResourceLink(rs, pagenum+1, pagesize)
	}
	return links
}

func GenBaseLink(rs JSONAPIResource) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.Trim(sinfo.GetBaseURL(), "/"),
		strings.Trim(sinfo.GetPrefix(), "/"),
	)
}

func GenSingleResourceLink(rs JSONAPIResource, id int64) string {
	return fmt.Sprintf(
		"%s/%s/%d",
		GenBaseLink(rs),
		rs.GetResourceName(),
		id,
	)
}

func GenMultiResourceLink(rs JSONAPIResource) string {
	return fmt.Sprintf(
		"%s/%s",
		GenBaseLink(rs),
		rs.GetResourceName(),
	)
}

func GenPaginatedResourceLink(rs JSONAPIResource, pagenum, pagesize int64) string {
	return fmt.Sprintf(
		"%s/%s?pagenum=%d&pagesize=%d",
		GenBaseLink(rs),
		rs.GetResourceName(),
		pagenum,
		pagesize,
	)
}

func GenSelfRelationshipLink(rs JSONAPIResource, rel string, id int64) string {
	return fmt.Sprintf(
		"%s/%s/%d/relationships/%s",
		GenBaseLink(rs),
		rs.GetResourceName(),
		id,
		rel,
	)
}

func GenRelatedRelationshipLink(rs JSONAPIResource, rel string, id int64) string {
	return fmt.Sprintf(
		"%s/%s/%d/%s",
		GenBaseLink(rs),
		rs.GetResourceName(),
		id,
		rel,
	)
}

//GetDefinedTagsWithValue check for fields that are initialized and returns a map
//with the tag and their values
func GetDefinedTagsWithValue(i interface{}, key string) map[string]interface{} {
	m := make(map[string]interface{})
	s := structs.New(i)
	for _, f := range s.Fields() {
		if !f.IsZero() {
			m[f.Tag(key)] = f.Value()
		}
	}
	return m
}

//GetDefinedTags check for fields that are initialized and returns a slice of
//their matching tag values
func GetDefinedTags(i interface{}, tag string) []string {
	var v []string
	s := structs.New(i)
	for _, f := range s.Fields() {
		if !f.IsZero() {
			v = append(v, f.Tag(tag))
		}
	}
	return v
}

// HandleCreateResponse modifies the grpc gateway filter which adds the JSON API header and
// modifies the http status response for POST request
func HandleCreateResponse(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if ok {
		trMD := md.TrailerMD
		if _, ok := trMD["method"]; ok {
			if trMD["method"][0] == "POST" {
				w.WriteHeader(http.StatusCreated)
			}
		}
	}
	return nil
}
