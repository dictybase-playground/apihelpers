// Package aphjsapi provides additional interfaces, wrapper and helper functions for original
// jsapi package("github.com/manyminds/api2go/jsapi")
package aphjsonapi

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dictyBase/go-middlewares/middlewares/pagination"
	jsapi "github.com/manyminds/api2go/jsonapi"
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

// AttributeToDbRowMapper is an interface to provide mapping between jsapi
// attribute and database row names. This is useful for implementing filter
// query parameter
type AttributeToDbRowMapper interface {
	GetMap() map[string]string
}

// RelationshipAttributes is an interface to provide attribute fields of
// relationship resources. This is mandatory for supporting sparse fieldset
// query parameter.
type RelationshipAttribute interface {
	GetAttributeFields(string) []string
}

// MarshalWithPagination adds pagination information for collection resource
func MarshalWithPagination(data interface{}, ep jsapi.ServerInformation, opt *pagination.Props) (*jsapi.Document, error) {
	var jst *jsapi.Document
	if reflect.TypeOf(data).Kind() != reflect.Slice {
		return jst, fmt.Errorf("%s\n", "Only slice type is allowed for pagination")
	}
	jst, err := MarshalToStructWrapper(data, ep)
	if err != nil {
		return jst, err
	}
	baseLink := jst.Links.Self
	pageLink := &jsapi.Links{}
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
// jsapi.Document structure for further json encoding
func MarshalToStructWrapper(data interface{}, ep jsapi.ServerInformation) (*jsapi.Document, error) {
	jst, err := jsapi.MarshalToStruct(data, ep)
	if err != nil {
		return jst, err
	}
	if len(jst.Data.DataArray) > 0 { //array resource objects
		// picking first element both from the generated and given typed structures
		elem := jst.Data.DataArray[0]
		value := reflect.ValueOf(data).Index(0).Interface()
		// link for the array resource itself
		jst.Links = &jsapi.Links{Self: generateMultiResourceLink(&elem, ep)}
		for i, d := range jst.Data.DataArray {
			// link for individual resource
			jst.Data.DataArray[i].Links = &jsapi.Links{Self: generateSingleResourceLink(&d, ep)}
			// Add relationships to every member
			r := generateRelationshipLinks(value, &d, ep)
			if len(r) > 0 {
				if len(jst.Included) > 0 && len(jst.Data.DataArray[i].Relationships) > 0 {
					for k, rel := range r {
						if excel, ok := jst.Data.DataArray[i].Relationships[k]; ok {
							excel.Links = rel.Links
							jst.Data.DataArray[i].Relationships[k] = excel
						} else {
							jst.Data.DataArray[i].Relationships[k] = rel
						}
					}
				} else {
					jst.Data.DataArray[i].Relationships = r
				}
			}
		}
		// Handle included members
		if len(jst.Included) > 0 {
			for i, m := range jst.Included {
				inrel := generateIncludedRelationshipLinks(value, &m, ep)
				if len(inrel) > 0 {
					jst.Included[i].Relationships = inrel
				}
			}
		}
	} else {
		// Handle included members
		if len(jst.Included) > 0 {
			for i, m := range jst.Included {
				inrel := generateIncludedRelationshipLinks(data, &m, ep)
				if len(inrel) > 0 {
					jst.Included[i].Relationships = inrel
				}
			}
		}
		jst.Links = &jsapi.Links{Self: generateSingleResourceLink(jst.Data.DataObject, ep)}
		relationships := generateRelationshipLinks(data, jst.Data.DataObject, ep)
		if len(relationships) > 0 {
			if len(jst.Included) > 0 && len(jst.Data.DataObject.Relationships) > 0 {
				for k, rel := range relationships {
					if excel, ok := jst.Data.DataObject.Relationships[k]; ok {
						excel.Links = rel.Links
						jst.Data.DataObject.Relationships[k] = excel
					} else {
						jst.Data.DataObject.Relationships[k] = rel
					}
				}
			} else {
				jst.Data.DataObject.Relationships = relationships
			}
		}
	}
	return jst, nil
}

func generateBaseLink(ep jsapi.ServerInformation) string {
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

func generateSingleResourceLink(jdata *jsapi.Data, ep jsapi.ServerInformation) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		generateBaseLink(ep),
		jdata.Type,
		jdata.ID,
	)
}

func generateMultiResourceLink(jdata *jsapi.Data, ep jsapi.ServerInformation) string {
	return fmt.Sprintf(
		"%s/%s",
		generateBaseLink(ep),
		jdata.Type,
	)
}

func generateRelationshipLinks(data interface{}, jdata *jsapi.Data, ep jsapi.ServerInformation) map[string]jsapi.Relationship {
	relationships := make(map[string]jsapi.Relationship)
	baselink := generateBaseLink(ep)
	self, ok := data.(MarshalSelfRelations)
	if ok {
		for _, rel := range self.GetSelfLinksInfo() {
			links := &jsapi.Links{}
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
			relationships[rel.Name] = jsapi.Relationship{Links: links}
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
				relationships[rel.Name] = jsapi.Relationship{Links: &jsapi.Links{Related: rlink}}
			}
		}
	}
	return relationships
}

func generateIncludedRelationshipLinks(data interface{}, jdata *jsapi.Data, ep jsapi.ServerInformation) map[string]jsapi.Relationship {
	relationships := make(map[string]jsapi.Relationship)
	if r, ok := data.(jsapi.MarshalIncludedRelations); ok {
		mi := r.GetReferencedStructs()
		if len(mi) > 0 {
			relationships = generateRelationshipLinks(mi[0], jdata, ep)
		}
	}
	return relationships
}

// MapFieldsToDbRow maps jsapi attributes to database row names
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

// GetTypeName gets the type name(type field) from a jsapi implementing
// interface. It is recommended to implement jsapi.EntityNamer interface to
// reduce the use of reflection
func GetTypeName(data interface{}) string {
	entity, ok := data.(jsapi.EntityNamer)
	if ok {
		return entity.GetName()
	}
	rType := reflect.TypeOf(data)
	if rType.Kind() == reflect.Ptr {
		return jsapi.Pluralize(jsapi.Jsonify(rType.Elem().Name()))
	}
	return jsapi.Pluralize(jsapi.Jsonify(rType.Name()))
}

// GetAttributeNames returns all JSAONAPI attribute names of data interface
func GetAttributeFields(data interface{}) []string {
	var attr []string
	t := reflect.TypeOf(data)
	if t == nil {
		return attr
	}
	var st reflect.Type
	if t.Kind() == reflect.Ptr {
		st = t.Elem()
	} else {
		st = t
	}
	for i := 0; i < st.NumField(); i++ {
		v, ok := st.Field(i).Tag.Lookup("json")
		if ok && v != "-" {
			attr = append(attr, v)
		}
	}
	return attr
}

// GetFilterAttributes gets all the JSONAPI attributes that are allowed to match filter query params
func GetFilterAttributes(data interface{}) []string {
	var attr []string
	t := reflect.TypeOf(data)
	if t == nil {
		return attr
	}
	var st reflect.Type
	if t.Kind() == reflect.Ptr {
		st = t.Elem()
	} else {
		st = t
	}
	for i := 0; i < st.NumField(); i++ {
		v, ok := st.Field(i).Tag.Lookup("json")
		if ok && v != "-" {
			_, ok := st.Field(i).Tag.Lookup("filter")
			if ok {
				attr = append(attr, v)
			}
		}
	}
	return attr
}

// GetAllRelationships returns all relationships of data interface
func GetAllRelationships(data interface{}) []RelationShipLink {
	var r []RelationShipLink
	self, ok := data.(MarshalSelfRelations)
	if ok {
		r = append(r, self.GetSelfLinksInfo()...)
	}
	related, ok := data.(MarshalRelatedRelations)
	if ok {
		r = append(r, related.GetRelatedLinksInfo()...)
	}
	return r
}

//GetRelatedTypes returns a map jsapi types of the related resources using
//reflection
func getRelatedTypeNames(data interface{}) []string {
	var names []string
	mtype := reflect.TypeOf((*jsapi.MarshalIdentifier)(nil)).Elem()
	t := reflect.TypeOf(data)
	for i := 0; i < t.NumField(); i++ {
		ftype := t.Field(i).Type
		if ftype.Kind() == reflect.Slice {
			if ftype.Elem().Implements(mtype) {
				names = append(
					names,
					jsapi.Pluralize(
						jsapi.Jsonify(
							ftype.Elem().Name(),
						),
					),
				)
				continue
			}
		}
		if ftype.Implements(mtype) {
			names = append(
				names,
				jsapi.Pluralize(
					jsapi.Jsonify(
						ftype.Name(),
					),
				),
			)
		}
	}
	return names
}
