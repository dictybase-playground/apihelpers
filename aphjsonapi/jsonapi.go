// Package aphjsonapi provides additional interfaces, wrapper and helper functions for original
// jsonapi package("github.com/manyminds/api2go/jsonapi")
package aphjsonapi

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dictyBase/go-middlewares/middlewares/pagination"
	"github.com/manyminds/api2go/jsonapi"
)

// RelationshipLink is a container type for having information about
// relationship links
type RelationShipLink struct {
	Name string `validate:"required"`
	// To override the default links, it will be appended to
	// the base url.
	SuffixFragment string
	// The type(type key in JSAONAPI specs) of relationship resource
	Type string `validate:"required"`
}

// MarshalSelfRelations is an interface  for creating self relationship links
type MarshalSelfRelations interface {
	// Validates all instances of RelationshipLink structures, using
	// https://gopkg.in/go-playground/validator.v9 package is recommended
	ValidateSelfLinks() error
	GetSelfLinksInfo() []RelationShipLink
}

// MarshalRelatedRelations is an interface  for creating related relationship
// links
type MarshalRelatedRelations interface {
	// Validates all instances of RelationshipLink structures, using
	// https://gopkg.in/go-playground/validator.v9 package is recommended
	ValidateRelatedLinks() error
	GetRelatedLinksInfo() []RelationShipLink
}

// AttributeToDbRowMapper is an interface to provide mapping between JSONAPI
// attribute and database row names
type AttributeToDbRowMapper interface {
	GetMap() map[string]string
}

// MarshalWithPagination adds pagination information for collection resource
func MarshalWithPagination(data interface{}, ep jsonapi.ServerInformation, opt *pagination.Props) (*jsonapi.Document, error) {
	var jst *jsonapi.Document
	if reflect.TypeOf(data).Kind() != reflect.Slice {
		return jst, fmt.Errorf("%s\n", "Only slice type is allowed for pagination")
	}
	jst, err := MarshalToStructWrapper(data, ep)
	if err != nil {
		return jst, err
	}
	baseLink := jst.Links.Self
	pageLink := &jsonapi.Links{}
	pageLink.Self = generatePaginatedResourceLink(baseLink, opt.Current, opt.Entries)
	pageLink.First = generatePaginatedResourceLink(baseLink, 1, opt.Entries)
	if opt.Current != 1 {
		pageLink.Previous = generatePaginatedResourceLink(baseLink, opt.Current-1, opt.Entries)
	}
	lastPage := int(math.Floor(float64(opt.Records) / float64(opt.Entries)))
	pageLink.Last = generatePaginatedResourceLink(baseLink, lastPage, opt.Entries)
	if opt.Current != lastPage {
		pageLink.Next = generatePaginatedResourceLink(baseLink, opt.Current+1, opt.Entries)
	}
	jst.Links = pageLink
	jst.Meta = map[string]interface{}{
		"pagination": map[string]int{
			"records": opt.Records,
			"total":   lastPage,
			"size":    opt.Entries,
			"number":  opt.Current,
		},
	}
	return jst, nil
}

// MarshalToStructWrapper adds relationship information and returns a
// jsonapi.Document structure for further json encoding
func MarshalToStructWrapper(data interface{}, ep jsonapi.ServerInformation) (*jsonapi.Document, error) {
	jst, err := jsonapi.MarshalToStruct(data, ep)
	if err != nil {
		return jst, err
	}
	if len(jst.Data.DataArray) > 0 { //array resource objects
		// picking first element both from the generated and given typed structures
		elem := jst.Data.DataArray[0]
		value := reflect.ValueOf(data).Index(0).Interface()
		// link for the array resource itself
		jst.Links = &jsonapi.Links{Self: generateMultiResourceLink(&elem, ep)}
		for i, d := range jst.Data.DataArray {
			// link for individual resource
			jst.Data.DataArray[i].Links = &jsonapi.Links{Self: generateSingleResourceLink(&d, ep)}
			// Add relationships to every member
			r := generateRelationshipLinks(value, &d, ep)
			jst.Data.DataArray[i].Relationships = r
		}
	} else {
		jst.Links = &jsonapi.Links{Self: generateSingleResourceLink(jst.Data.DataObject, ep)}
		relationships := generateRelationshipLinks(data, jst.Data.DataObject, ep)
		if len(relationships) > 0 {
			jst.Data.DataObject.Relationships = relationships
		}
	}
	return jst, nil
}

func generateBaseLink(ep jsonapi.ServerInformation) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.Trim(ep.GetBaseURL(), "/"),
		strings.Trim(ep.GetPrefix(), "/"),
	)
}

func generatePaginatedResourceLink(baseurl string, pagenum, pagesize int) string {
	return fmt.Sprintf(
		"%s?page[number]=%d&page[size]=%d",
		baseurl,
		pagenum,
		pagesize,
	)
}

func generateSingleResourceLink(jdata *jsonapi.Data, ep jsonapi.ServerInformation) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		generateBaseLink(ep),
		jdata.Type,
		jdata.ID,
	)
}

func generateMultiResourceLink(jdata *jsonapi.Data, ep jsonapi.ServerInformation) string {
	return fmt.Sprintf(
		"%s/%s",
		generateBaseLink(ep),
		jdata.Type,
	)
}

func generateRelationshipLinks(data interface{}, jdata *jsonapi.Data, ep jsonapi.ServerInformation) map[string]jsonapi.Relationship {
	relationships := make(map[string]jsonapi.Relationship)
	baselink := generateBaseLink(ep)
	self, ok := data.(MarshalSelfRelations)
	if ok {
		for _, rel := range self.GetSelfLinksInfo() {
			links := &jsonapi.Links{}
			if len(rel.SuffixFragment) > 0 {
				links.Self = fmt.Sprintf("%s/%s", baselink, strings.Trim(rel.SuffixFragment, "/"))
			} else {
				links.Self = fmt.Sprintf("%s/%s/%s/relationships/%s",
					baselink,
					jdata.Type,
					jdata.ID,
					rel.Name,
				)
			}
			relationships[rel.Name] = jsonapi.Relationship{Links: links}
		}
	}
	related, ok := data.(MarshalRelatedRelations)
	if ok {
		for _, rel := range related.GetRelatedLinksInfo() {
			var rlink string
			if len(rel.SuffixFragment) > 0 {
				rlink = fmt.Sprintf("%s/%s", baselink, strings.Trim(rel.SuffixFragment, "/"))
			} else {
				rlink = fmt.Sprintf("%s/%s/%s/%s",
					baselink,
					jdata.Type,
					jdata.ID, rel.Name,
				)
			}
			if _, ok := relationships[rel.Name]; ok {
				relationships[rel.Name].Links.Related = rlink
			} else {
				relationships[rel.Name] = jsonapi.Relationship{Links: &jsonapi.Links{Related: rlink}}
			}
		}
	}
	return relationships
}

// MapFieldsToDbRow maps JSONAPI attributes to database row names
func MapFieldsToDbRow(data interface{}) map[string]string {
	m, ok := data.(AttributeToDbRowMapper)
	if ok {
		return m.GetMap()
	}
	frow := make(map[string]string)
	t := reflect.TypeOf(data)
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		v, ok := tag.Lookup("json")
		if ok && v != "-" {
			r, dbok := tag.Lookup("db")
			if dbok && r != "-" {
				frow[v] = r
			}
		}
	}
	return frow
}
