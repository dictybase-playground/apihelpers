package aphgrpc

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"gopkg.in/mgutz/dat.v2/dat"
	"gopkg.in/mgutz/dat.v2/sqlx-runner"

	"github.com/dictyBase/go-genproto/dictybaseapis/api/jsonapi"
	"github.com/fatih/structs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/metadata"
)

const (
	DefaultPagenum  int64 = 1
	DefaultPagesize int64 = 10
)

type contextKey string

// String output the details of context key
func (c contextKey) String() string {
	return "pagination context key " + string(c)
}

var (
	ContextKeyParams  = contextKey("params")
	ContextKeyInclude = contextKey("includeStr")
	ContextKeyFilter  = contextKey("filterStr")
	ContextKeyFields  = contextKey("fieldsStr")
	ContextKeyIsList  = contextKey("isListMethod")
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

func NullToTime(nt dat.NullTime) *timestamp.Timestamp {
	var ts *timestamp.Timestamp
	if nt.Valid {
		ts, _ := ptypes.TimestampProto(nt.Time)
		return ts
	}
	return ts
}

func ProtoTimeStamp(ts *timestamp.Timestamp) time.Time {
	t, _ := ptypes.Timestamp(ts)
	return t
}

func TimestampProto(t time.Time) *timestamp.Timestamp {
	ts, _ := ptypes.TimestampProto(t)
	return ts
}

func NullToString(s dat.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func NullToInt64(i dat.NullInt64) int64 {
	if i.Valid {
		return i.Int64
	}
	var i64 int64
	return i64
}

// GetTotalPageNum calculate total no of pages from total no. records and page size
func GetTotalPageNum(record, pagesize int64) int64 {
	total := int64(math.Floor(float64(record) / float64(pagesize)))
	if math.Mod(float64(record), float64(pagesize)) > 0 {
		total += 1
	}
	return total
}

// GenPaginatedLinks generates paginated resource links
// from various page properties.
func GenPaginatedLinks(url string, lastpage, pagenum, pagesize int64) map[string]string {
	links := make(map[string]string)
	links["self"] = AppendPaginationParams(url, pagenum, pagesize)
	links["first"] = AppendPaginationParams(url, 1, pagesize)
	if pagenum != 1 {
		links["previous"] = AppendPaginationParams(url, pagenum-1, pagesize)
	}
	links["last"] = AppendPaginationParams(url, lastpage, pagesize)
	if pagenum != lastpage {
		links["next"] = AppendPaginationParams(url, pagenum+1, pagesize)
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
		"%s/%d",
		GenBaseLink(rs),
		id,
	)
}

func GenMultiResourceLink(rs JSONAPIResource) string {
	return GenBaseLink(rs)
}

func AppendPaginationParams(url string, pagenum, pagesize int64) string {
	return fmt.Sprintf("%s?pagenum=%d&pagesize=%d", url, pagenum, pagesize)
}

func GenPaginatedResourceLink(rs JSONAPIResource, pagenum, pagesize int64) string {
	return fmt.Sprintf(
		"%s?pagenum=%d&pagesize=%d",
		GenBaseLink(rs),
		pagenum,
		pagesize,
	)
}

func GenSelfRelationshipLink(rs JSONAPIResource, rel string, id int64) string {
	return fmt.Sprintf(
		"%s/%d/relationships/%s",
		GenBaseLink(rs),
		id,
		rel,
	)
}

func GenRelatedRelationshipLink(rs JSONAPIResource, rel string, id int64) string {
	return fmt.Sprintf(
		"%s/%d/%s",
		GenBaseLink(rs),
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
			switch trMD["method"][0] {
			case "POST":
				w.WriteHeader(http.StatusCreated)
			case "POST_NO_CONTENT":
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
	return nil
}

// SkipHTTPLinks looks up the context for the presence of gprc metadata
// for skipping the generation of HTTP links
func SkipHTTPLinks(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	if _, ok := md["skip-http-links"]; ok {
		return true
	}
	return false
}

// ListReqCtx generate context data from list(collection) request
func ListReqCtx(params *JSONAPIParams, r *jsonapi.ListRequest) context.Context {
	ctx := context.WithValue(context.Background(), ContextKeyIsList, "yes")
	ctx = context.WithValue(ctx, ContextKeyParams, params)
	if params.HasInclude {
		ctx = context.WithValue(ctx, ContextKeyInclude, r.Include)
	}
	if params.HasFields {
		ctx = context.WithValue(ctx, ContextKeyFields, r.Fields)
	}
	if params.HasFilter {
		ctx = context.WithValue(ctx, ContextKeyFilter, r.Filter)
	}
	return ctx
}

// GetReqCtx generate a context data from get request
func GetReqCtx(params *JSONAPIParams, r *jsonapi.GetRequest) context.Context {
	ctx := context.WithValue(context.Background(), ContextKeyParams, params)
	if params.HasInclude {
		ctx = context.WithValue(ctx, ContextKeyInclude, r.Include)
	}
	if params.HasFields {
		ctx = context.WithValue(ctx, ContextKeyFields, r.Fields)
	}
	return ctx
}

// AssignFieldsToStructs copy fields value
// between structure
func AssignFieldsToStructs(from interface{}, to interface{}) {
	toR := structs.New(to)
	for _, f := range structs.New(from).Fields() {
		if !f.IsZero() {
			if nf, ok := toR.FieldOk(f.Name()); ok {
				nf.Set(f.Value())
			}
		}
	}
}

type ServiceOptions struct {
	PathPrefix      string
	Resource        string
	BaseURL         string
	Include         []string
	FieldsToColumns map[string]string
	FilToColumns    map[string]string
	ReqAttrs        []string
	Topics          map[string]string
}

type Option func(*ServiceOptions)

func TopicsOption(t map[string]string) Option {
	return func(so *ServiceOptions) {
		so.Topics = t
	}
}

func JSONAPIResourceOptions(prefix, resource, base string) Option {
	return func(so *ServiceOptions) {
		so.PathPrefix = prefix
		so.Resource = resource
		so.BaseURL = base
	}
}

func BaseURLOption(base string) Option {
	return func(so *ServiceOptions) {
		so.BaseURL = base
	}
}

func ReqAttributesOption(attr []string) Option {
	return func(so *ServiceOptions) {
		so.ReqAttrs = attr
	}
}

func FilterMappingOptions(fmap map[string]string) Option {
	return func(so *ServiceOptions) {
		so.FilToColumns = fmap
	}
}

func FieldsMappingOptions(fmap map[string]string) Option {
	return func(so *ServiceOptions) {
		so.FieldsToColumns = fmap
	}
}

func IncludeOptions(inc []string) Option {
	return func(so *ServiceOptions) {
		so.Include = inc
	}
}

type Service struct {
	Dbh             *runner.DB
	PathPrefix      string
	Include         []string
	FieldsToColumns map[string]string
	Resource        string
	BaseURL         string
	FilToColumns    map[string]string
	ListMethod      bool
	ReqAttrs        []string
	Context         context.Context
	Topics          map[string]string
}

func (s *Service) RequiredAttrs() []string {
	return s.ReqAttrs
}

func (s *Service) IsListMethod() bool {
	return s.ListMethod
}

func (s *Service) FilterToColumns() map[string]string {
	return s.FilToColumns
}

func (s *Service) AllowedFilter() []string {
	var f []string
	for k, _ := range s.FilterToColumns() {
		f = append(f, k)
	}
	return f
}

func (s *Service) AllowedInclude() []string {
	return s.Include
}

func (s *Service) AllowedFields() []string {
	var f []string
	for k, _ := range s.FieldsToColumns {
		f = append(f, k)
	}
	return f
}

func (s *Service) GetResourceName() string {
	return s.Resource
}

func (s *Service) GetBaseURL() string {
	return s.BaseURL
}

func (s *Service) GetPathPrefix() string {
	return s.PathPrefix
}

func (s *Service) MapFieldsToColumns(fields []string) []string {
	var columns []string
	for _, v := range fields {
		columns = append(columns, s.FieldsToColumns[v])
	}
	return columns
}

func (s *Service) GetCount(ctx context.Context, table string) (int64, error) {
	var count int64
	err := s.Dbh.Select("COUNT(*)").From(table).QueryScalar(&count)
	return count, err
}

func (s *Service) GetAllFilteredCount(ctx context.Context, table string) (int64, error) {
	var count int64
	params, ok := ctx.Value(ContextKeyParams).(*JSONAPIParams)
	if !ok {
		return count, fmt.Errorf("no params object found in context")
	}
	err := s.Dbh.Select("COUNT(*)").
		From(table).
		Scope(
			FilterToWhereClause(s, params.Filters),
			FilterToBindValue(params.Filters)...,
		).QueryScalar(&count)
	return count, err
}

// GetRelatedPagination generates JSONAPI pagination links for relation resources
func (s *Service) GetRelatedPagination(id, record, pagenum, pagesize int64, relation string) (*jsonapi.PaginationLinks, int64) {
	pages := GetTotalPageNum(record, pagesize)
	baseLink := s.GenCollResourceRelSelfLink(id, relation)
	pageLinks := GenPaginatedLinks(baseLink, pages, pagenum, pagesize)
	jsapiLinks := &jsonapi.PaginationLinks{
		Self:  pageLinks["self"],
		Last:  pageLinks["last"],
		First: pageLinks["first"],
	}
	if _, ok := pageLinks["previous"]; ok {
		jsapiLinks.Prev = pageLinks["previous"]
	}
	if _, ok := pageLinks["next"]; ok {
		jsapiLinks.Next = pageLinks["next"]
	}
	return jsapiLinks, pages
}

// GetPagination generates JSONAPI pagination links along with fields, include and filter query parameters
func (s *Service) GetPagination(ctx context.Context, record, pagenum, pagesize int64) (*jsonapi.PaginationLinks, int64) {
	pages := GetTotalPageNum(record, pagesize)
	baseLink := s.GenCollResourceSelfLink(ctx)
	pageLinks := GenPaginatedLinks(baseLink, pages, pagenum, pagesize)
	pageType := []string{"self", "last", "first", "previous", "next"}
	params, ok := ctx.Value(ContextKeyParams).(*JSONAPIParams)
	includeStr, _ := ctx.Value(ContextKeyInclude).(string)
	fieldsStr, _ := ctx.Value(ContextKeyFields).(string)
	filterStr, _ := ctx.Value(ContextKeyFilter).(string)
	if ok {
		switch {
		case params.HasFields && params.HasInclude && params.HasFilter:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&fields=%s&include=%s&filter=%s", fieldsStr, includeStr, filterStr)
				}
			}
		case params.HasFields && params.HasInclude:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&fields=%s&include=%s", fieldsStr, includeStr)
				}
			}
		case params.HasFields && params.HasFilter:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&fields=%s&filter=%s", fieldsStr, filterStr)
				}
			}
		case params.HasInclude && params.HasFilter:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&include=%s&filter=%s", includeStr, filterStr)
				}
			}
		case params.HasInclude:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&include=%s", includeStr)
				}
			}
		case params.HasFilter:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&filter=%s", filterStr)
				}
			}
		case params.HasFields:
			for _, v := range pageType {
				if _, ok := pageLinks[v]; ok {
					pageLinks[v] += fmt.Sprintf("&fields=%s", fieldsStr)
				}
			}
		}
	}
	jsapiLinks := &jsonapi.PaginationLinks{
		Self:  pageLinks["self"],
		Last:  pageLinks["last"],
		First: pageLinks["first"],
	}
	if _, ok := pageLinks["previous"]; ok {
		jsapiLinks.Prev = pageLinks["previous"]
	}
	if _, ok := pageLinks["next"]; ok {
		jsapiLinks.Next = pageLinks["next"]
	}
	return jsapiLinks, pages
}

func (s *Service) GenCollResourceRelSelfLink(id int64, relation string) string {
	return fmt.Sprintf(
		"%s/%d/%s",
		GenMultiResourceLink(s),
		id,
		relation,
	)
}

func (s *Service) GenCollResourceSelfLink(ctx context.Context) string {
	link := GenMultiResourceLink(s)
	params, ok := ctx.Value(ContextKeyParams).(*JSONAPIParams)
	if !ok {
		return link
	}
	includeStr, _ := ctx.Value(ContextKeyInclude).(string)
	fieldsStr, _ := ctx.Value(ContextKeyFields).(string)
	filterStr, _ := ctx.Value(ContextKeyFilter).(string)
	switch {
	case params.HasFields && params.HasFilter && params.HasInclude:
		link += fmt.Sprintf("?fields=%s&include=%s&filter=%s", fieldsStr, includeStr, filterStr)
	case params.HasFields && params.HasFilter:
		link += fmt.Sprintf("?fields=%s&filter=%s", fieldsStr, filterStr)
	case params.HasFields && params.HasInclude:
		link += fmt.Sprintf("?fields=%s&include=%s", fieldsStr, includeStr)
	case params.HasFilter && params.HasInclude:
		link += fmt.Sprintf("?filter=%s&include=%s", filterStr, includeStr)
	case params.HasInclude:
		link += fmt.Sprintf("?include=%s", includeStr)
	case params.HasFilter:
		link += fmt.Sprintf("?filter=%s", filterStr)
	case params.HasFields:
		link += fmt.Sprintf("?fields=%s", fieldsStr)
	}
	return link
}

func (s *Service) GenResourceSelfLink(ctx context.Context, id int64) string {
	links := GenSingleResourceLink(s, id)
	params, ok := ctx.Value(ContextKeyParams).(*JSONAPIParams)
	_, lok := ctx.Value(ContextKeyIsList).(string)
	if !lok && ok {
		includeStr, _ := ctx.Value(ContextKeyInclude).(string)
		fieldsStr, _ := ctx.Value(ContextKeyFields).(string)
		switch {
		case params.HasFields && params.HasInclude:
			links += fmt.Sprintf("?fields=%s&include=%s", fieldsStr, includeStr)
		case params.HasFields:
			links += fmt.Sprintf("?fields=%s", fieldsStr)
		case params.HasInclude:
			links += fmt.Sprintf("?include=%s", includeStr)
		}
	}
	return links
}

func (s *Service) SetBaseURL() error {
	md, ok := metadata.FromIncomingContext(s.Context)
	if !ok {
		return ErrRetrieveMetadata
	}
	if slice, ok := md["x-forwarded-host"]; ok {
		s.BaseURL = fmt.Sprintf("http://%s", slice[0])
	}
	return nil
}
