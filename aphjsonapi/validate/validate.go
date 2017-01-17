package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dictyBase/apihelpers/aphcollection"
	jsapi "github.com/dictyBase/apihelpers/aphjsonapi"
	"github.com/manyminds/api2go/jsonapi"
)

//HasRelationships matches if the relationships rels are implemented in the
//given JSONAPI implementing data
func HasRelationships(data interface{}, rels []string) error {
	matchedRel := true
	matchedSelf := true
	self, ok := data.(jsapi.MarshalSelfRelations)
	if ok {
		for _, rel := range self.GetSelfLinksInfo() {
			if aphcollection.Contains(rels, rel.Name) {
				matchedSelf = true
				break
			}
		}
	}
	related, ok := data.(jsapi.MarshalRelatedRelations)
	if ok {
		for _, rel := range related.GetRelatedLinksInfo() {
			if !aphcollection.Contains(rels, rel.Name) {
				matchedRel = true
			}
		}
	}
	if !matchedRel && !matchedSelf {
		return fmt.Errorf("given names %s does not matches to any relationship", strings.Join(rels, ","))
	}
	return nil
}

// ResourceType matches name with JSONAPI type in data's fields
func ResourceType(name string, data interface{}) bool {
	if name == getTypeName(data) {
		return true
	}
	return false
}

// RelatedResourceType matches name with all related JSONAPI type in data's fields
func RelatedResourceType(name string, data interface{}) bool {
	if aphcollection.Contains(GetRelatedTypeNames(data), name) {
		return true
	}
	return false
}

// FieldNames matches all elements of s with all JSONAPI field names
// in data's fields
func FieldNames(s []string, data interface{}) bool {
	t := reflect.TypeOf(data)
	for i := 0; i < t.NumField(); i++ {
		v, ok := t.Field(i).Tag.Lookup("json")
		if ok && v != "-" && !aphcollection.Contains(s, v) {
			return false
		}
	}
	return true
}

//GetRelatedTypeNames returns the JSONAPI types of the related resources
func GetRelatedTypeNames(data interface{}) []string {
	var names []string
	mtype := reflect.TypeOf((*jsonapi.MarshalIdentifier)(nil)).Elem()
	t := reflect.TypeOf(data)
	for i := 0; i < t.NumField(); i++ {
		ftype := t.Field(i).Type
		if ftype.Kind() == reflect.Slice {
			if ftype.Elem().Implements(mtype) {
				names = append(
					names,
					jsonapi.Pluralize(
						jsonapi.Jsonify(
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
				jsonapi.Pluralize(
					jsonapi.Jsonify(
						ftype.Name(),
					),
				),
			)
		}
	}
	return names
}

func getTypeName(data interface{}) string {
	entity, ok := data.(jsonapi.EntityNamer)
	if ok {
		return entity.GetName()
	}
	rType := reflect.TypeOf(data)
	if rType.Kind() == reflect.Ptr {
		return jsonapi.Pluralize(jsonapi.Jsonify(rType.Elem().Name()))
	}
	return jsonapi.Pluralize(jsonapi.Jsonify(rType.Name()))
}
