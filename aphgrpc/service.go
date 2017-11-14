package aphgrpc

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"gopkg.in/mgutz/dat.v1/sqlx-runner"

	"github.com/dictyBase/apihelpers/aphgrpc"
	"github.com/dictyBase/go-genproto/dictybaseapis/api/jsonapi"
	"github.com/fatih/structs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
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
		strings.Trim(rs.GetBaseURL(), "/"),
		strings.Trim(rs.GetPathPrefix(), "/"),
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

// ConvertAllToAny generates slice of arbitrary serialized protocol buffer
// message
func ConvertAllToAny(msg []proto.Message) ([]*any.Any, error) {
	as := make([]*any.Any, len(msg))
	for i, p := range msg {
		pkg, err := ptypes.MarshalAny(p)
		if err != nil {
			return as, err
		}
		as[i] = pkg
	}
	return as, nil
}

type Service struct {
	Dbh             *runner.DB
	pathPrefix      string
	include         []string
	includeStr      string
	fieldsToColumns map[string]string
	fieldsStr       string
	resource        string
	baseURL         string
	filterToColumns map[string]string
	filterStr       string
	params          *JSONAPIParams
	listMethod      bool
	requiredAttrs   []string
}

func (s *Service) RequiredAttrs() []string {
	return s.requiredAttrs
}

func (s *Service) IsListMethod() bool {
	return s.listMethod
}

func (s *Service) FilterToColumns() map[string]string {
	return s.filterToColumns
}

func (s *Service) AllowedFilter() []string {
	var f []string
	for k, _ := range s.filterToColumns {
		f = append(f, k)
	}
	return f
}

func (s *Service) AllowedInclude() []string {
	return s.include
}

func (s *Service) AllowedFields() []string {
	var f []string
	for k, _ := range s.fieldsToColumns {
		f = append(f, k)
	}
	return f
}

func (s *Service) GetResourceName() string {
	return s.resource
}

func (s *Service) GetBaseURL() string {
	return s.baseURL
}

func (s *Service) GetPathPrefix() string {
	return s.pathPrefix
}

func (s *Service) MapFieldsToColumns(fields []string) []string {
	var columns []string
	for _, v := range fields {
		columns = append(columns, s.fieldsToColumns[v])
	}
	return columns
}

func (s *Service) getCount(table string) (int64, error) {
	var count int64
	err := s.Dbh.Select("COUNT(*)").From(table).QueryScalar(&count)
	return count, err
}

func (s *Service) getAllFilteredCount(table string) (int64, error) {
	var count int64
	err := s.Dbh.Select("COUNT(*)").
		From(table).
		Scope(
			aphgrpc.FilterToWhereClause(s, s.params.Filter),
			aphgrpc.FilterToBindValue(s.params.Filter)...,
		).QueryScalar(&count)
	return count, err
}

func (s *Service) getPagination(record, pagenum, pagesize int64) (*jsonapi.PaginationLinks, int64) {
	pages := GetTotalPageNum(record, pagenum, pagesize)
	pageLinks := GetPaginatedLinks(s, pages, pagenum, pagesize)
	pageType := []string{"self", "last", "first", "previous", "next"}
	params := s.params
	switch {
	case params.HasFields && params.HasInclude && params.HasFilter:
		for _, v := range pageType {
			if _, ok := pageLinks[v]; ok {
				pageLinks[v] += fmt.Sprintf("%s&fields=%s&include=%s&filter=%s", s.fieldsStr, s.includeStr, s.filterStr)
			}
		}
	case params.HasFields && params.HasInclude:
		for _, v := range pageType {
			if _, ok := pageLinks[v]; ok {
				pageLinks[v] += fmt.Sprintf("%s&fields=%s&include=%s", s.fieldsStr, s.includeStr)
			}
		}
	case params.HasFields && params.HasFilter:
		for _, v := range pageType {
			if _, ok := pageLinks[v]; ok {
				pageLinks[v] += fmt.Sprintf("%s&fields=%s&filter=%s", s.fieldsStr, s.filterStr)
			}
		}
	case params.HasInclude && params.HasFilter:
		for _, v := range pageType {
			if _, ok := pageLinks[v]; ok {
				pageLinks[v] += fmt.Sprintf("%s&include=%s&filter=%s", s.includeStr, s.filterStr)
			}
		}
	}
	jsapiLinks := jsonapi.PaginationLinks{
		Self:  pageLinks["self"],
		Last:  pageLinks["last"],
		First: pageLinks["first"],
	}
	if _, ok := pageLinks["previous"]; ok {
		jsapiLinks.Previous = pageLinks["previous"]
	}
	if _, ok := pageLinks["next"]; ok {
		jsapiLinks.Next = pageLinks["next"]
	}
	return jsapiLinks, pages
}

func (s *Service) genCollResourceSelfLink() string {
	link := GenMultiResourceLink(s)
	params := s.params
	switch {
	case params.HasFields && params.HasFilter && params.HasInclude:
		link += fmt.Sprintf("?fields=%s&include=%s&filter=%s", s.fieldsStr, s.includeStr, s.filterStr)
	case params.HasFields && params.HasFilter:
		link += fmt.Sprintf("?fields=%s&filter=%s", s.fieldsStr, s.filterStr)
	case params.HasFields && params.HasInclude:
		link += fmt.Sprintf("?fields=%s&include=%s", s.fieldsStr, s.includeStr)
	case params.HasFilter && params.HasInclude:
		link += fmt.Sprintf("?filter=%s&include=%s", s.filterStr, s.includeStr)
	}
	return link
}

func (s *Service) genResourceSelfLink(id int64) string {
	links := GenSingleResourceLink(s, id)
	if !s.IsListMethod() && s.params != nil {
		params := s.params
		switch {
		case params.HasFields && params.HasIncludes:
			links += fmt.Sprintf("?fields=%s&include=%s", s.fieldsStr, s.includeStr)
		case params.HasFields:
			links += fmt.Sprintf("?fields=%s", s.fieldsStr)
		case params.HasIncludes:
			links += fmt.Sprintf("?include=%s", s.includeStr)
		}
	}
	return links
}
