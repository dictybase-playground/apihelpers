package validate

import (
	"fmt"
	"reflect"

	"github.com/dictyBase/apihelpers/aphcollection"
	jsapi "github.com/dictyBase/apihelpers/aphjsonapi"
	"github.com/manyminds/api2go/jsonapi"
)

func mapRelsToName(js []jsapi.RelationShipLink, fn func(jsapi.RelationShipLink) string) []string {
	s := make([]string, len(js))
	for i, v := range js {
		s[i] = fn(v)
	}
	return s
}

// GetAllRelationships returns all relationships of data interface
func GetAllRelationships(data interface{}) ([]jsapi.RelationShipLink, error) {
	var r []jsapi.RelationShipLink
	self, ok := data.(jsapi.MarshalSelfRelations)
	if ok {
		if err := self.ValidateSelfLinks(); err != nil {
			return r, err
		}
		r = append(r, self.GetSelfLinksInfo()...)
	}
	related, ok := data.(jsapi.MarshalRelatedRelations)
	if ok {
		if err := related.ValidateRelatedLinks(); err != nil {
			return r, err
		}
		r = append(r, related.GetRelatedLinksInfo()...)
	}
	if len(r) == 0 {
		return r, fmt.Errorf("no relationship defined")
	}
	return r, nil
}

//HasRelationships checks if slice a contains any relationship name in rs slice
func HasRelationships(a []string, rs []jsapi.RelationShipLink) error {
	allNames := mapRelsToName(rs, func(r jsapi.RelationShipLink) string {
		return r.Name
	})
	for _, s := range a {
		if !aphcollection.Contains(allNames, s) {
			return fmt.Errorf(
				"given name %s does not matches to any defined relationships",
				s,
			)
		}
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

// RelationshipResourceType checks if the given resource name matches any type
// in the rs slice
func RelationshipResourceType(name string, rs []jsapi.RelationShipLink) error {
	for _, r := range rs {
		if name == r.Type {
			return nil
		}
	}
	return fmt.Errorf("%s resource type does not matches any relationship type", name)
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
