package aphgrpc

import (
	"fmt"
	"math"
	"strings"
)

const (
	DefaultPagenum  = 1
	DefaultPagesize = 10
)

// JSONAPIAllowedParams interface should be implement by all grpc-gateway services
// that supports JSON API specifications.
type JSONAPIAllowedParams interface {
	// Relationships that could be included
	AllowedInclude() []string
	// Attribute fields that are allowed
	AllowedFields() []string
	// Filter fields that are allowed
	AllowedFilter() []string
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

// GetPaginatedLinks generates links with pagination for collection resources
func GetPaginatedLinks(rs JSONAPIResource, total, pagenum, pagesize int64) map[string]string {
	var links map[string]string
	links["self"] = GenPaginatedResourceLink(rs, pagenum, pagesize)
	links["first"] = GenPaginatedResourceLink(rs, 1, pagesize)
	if pagenum != 1 {
		links["previous"] = GenPaginatedResourceLink(rs, pagenum-1, pagesize)
	}
	lastPage := int(math.Floor(float64(total) / float64(pagesize)))
	if math.Mod(float64(total), float64(pagesize)) > 0 {
		lastPage += 1
	}
	links["last"] = GenPaginatedResourceLink(rs, lastPage, pagesize)
	if pagenum != lastPage {
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
