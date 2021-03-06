package validate

import (
	"fmt"
	"reflect"

	"gopkg.in/go-playground/validator.v9"

	"github.com/dictyBase/apihelpers/aphcollection"
	jsapi "github.com/dictyBase/apihelpers/aphjsonapi"
	"github.com/manyminds/api2go/jsonapi"
)

var validate = validator.New()

func mapRelsToName(js []jsapi.RelationShipLink, fn func(jsapi.RelationShipLink) string) []string {
	s := make([]string, len(js))
	for i, v := range js {
		s[i] = fn(v)
	}
	return s
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

// RelationshipResourceType checks if the given resource name matches any type
// in the relationship rs slice and returns the type name
func RelationshipResourceType(name string, data interface{}) (string, error) {
	if japi, ok := data.(jsonapi.MarshalReferences); ok {
		for _, r := range japi.GetReferences() {
			if name == r.Type {
				return r.Type, nil
			}
		}
	}
	return "", fmt.Errorf("%s resource type does not matches any relationship type", name)
}

//getRelatedTypeNames returns the JSONAPI types of the related resources using
//reflection
func getRelatedTypeNames(data interface{}) []string {
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
