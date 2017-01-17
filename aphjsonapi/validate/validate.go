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
