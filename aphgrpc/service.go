package aphgrpc

import (
	"fmt"
	"strings"
)

// JSONAPIAllowedParams interface should be implement by all grpc-gateway services
// that supports JSON API specifications.
type JSONAPIAllowedParams interface {
	// Relationships that could be included
	AllowedInclude() []string
	// Attribute fields that are allowed
	AllowedFields() []string
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

func GenPaginatedResourceLink(rs JSONAPIResource) string {
	return ""
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
