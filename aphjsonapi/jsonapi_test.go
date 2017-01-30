package aphjsonapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/dictyBase/apihelpers/aphcollection"
	"github.com/dictyBase/apihelpers/aphtest"
	"github.com/dictyBase/go-middlewares/middlewares/pagination"
	jsapi "github.com/manyminds/api2go/jsonapi"
)

func TestLinkGeneration(t *testing.T) {
	srvinfo := aphtest.NewTestApiInfo()
	exlink := fmt.Sprintf("%s/%s", aphtest.APIHost, aphtest.PathPrefix)
	bslink := generateBaseLink(srvinfo)
	if bslink != exlink {
		t.Fatalf("expected base link %s does not match with generated link %s", exlink, bslink)
	}

	id := "14"
	jtype := "books"
	exrlink := fmt.Sprintf("%s/%s/%s", exlink, jtype, id)
	rlink := generateSingleResourceLink(&jsapi.Data{Type: jtype, ID: id}, srvinfo)
	if exrlink != rlink {
		t.Fatalf("expected single resource link %s does not match generated link %s", exrlink, rlink)
	}
}

func TestSelfLink(t *testing.T) {
	perm := &Permission{
		ID:          "10",
		Permission:  "gene curation",
		Description: "Authority to edit gene information",
	}
	srvinfo := aphtest.NewTestApiInfo()
	pstruct, err := MarshalToStructWrapper(perm, srvinfo)
	if err != nil {
		t.Errorf("error in marshaling to structure %s\n", err)
	}
	exstruct := &jsapi.Document{
		Links: &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s/%s", srvinfo.GetBaseURL(), srvinfo.GetPrefix(), "permissions", "10")},
		Data: &jsapi.DataContainer{
			DataObject: &jsapi.Data{
				Type:       "permissions",
				ID:         "10",
				Attributes: []byte(`{"permission":"gene curation","description":"Authority to edit gene information"}`),
			},
		},
	}
	if !reflect.DeepEqual(pstruct, exstruct) {
		t.Fatal("expected and generated jsonapi structure did not match")
	}
}

func TestRelationshipLink(t *testing.T) {
	srvinfo := aphtest.NewTestApiInfo()
	bslink := generateBaseLink(srvinfo)
	u := &User{
		ID:    "32",
		Name:  "Tucker",
		Email: "tucker@jumbo.com",
	}
	pstruct, err := MarshalToStructWrapper(u, srvinfo)
	if err != nil {
		t.Errorf("error in marshaling to structure %s\n", err)
	}
	rel := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "users", "32", "roles"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "users", "32", "roles"),
		},
	}
	exstruct := &jsapi.Document{
		Links: &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s", bslink, "users", "32")},
		Data: &jsapi.DataContainer{
			DataObject: &jsapi.Data{
				Type:          "users",
				ID:            "32",
				Attributes:    []byte(`{"name":"Tucker","email":"tucker@jumbo.com"}`),
				Relationships: map[string]jsapi.Relationship{"roles": rel},
			},
		},
	}
	if !reflect.DeepEqual(pstruct, exstruct) {
		t.Fatal("expected and generated jsonapi structure did not match")
	}
}

func TestMultiRelationshipsLink(t *testing.T) {
	srvinfo := aphtest.NewTestApiInfo()
	bslink := generateBaseLink(srvinfo)
	r := &Role{
		ID:          "44",
		Role:        "Administrator",
		Description: "The God",
	}
	pstruct, err := MarshalToStructWrapper(r, srvinfo)
	if err != nil {
		t.Errorf("error in marshaling to structure %s\n", err)
	}
	prel := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "roles", "44", "permissions"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "roles", "44", "permissions"),
		},
	}
	urel := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "roles", "44", "users"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "roles", "44", "consumers"),
		},
	}
	exstruct := &jsapi.Document{
		Links: &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s", bslink, "roles", "44")},
		Data: &jsapi.DataContainer{
			DataObject: &jsapi.Data{
				Type:          "roles",
				ID:            "44",
				Attributes:    []byte(`{"role":"Administrator","description":"The God"}`),
				Relationships: map[string]jsapi.Relationship{"permissions": prel, "users": urel},
			},
		},
	}
	if !reflect.DeepEqual(pstruct, exstruct) {
		t.Fatal("expected and generated jsonapi structure did not match")
	}
}

func TestCollectionRelationshipsLink(t *testing.T) {
	srvinfo := aphtest.NewTestApiInfo()
	bslink := generateBaseLink(srvinfo)
	users := []*User{
		&User{
			ID:    "12",
			Name:  "Caboose",
			Email: "caboose@caboose.com",
		},
		&User{
			ID:    "21",
			Name:  "Damon",
			Email: "damon@damon.com",
		},
	}
	pstruct, err := MarshalToStructWrapper(users, srvinfo)
	if err != nil {
		t.Errorf("error in marshaling to structure %s\n", err)
	}
	rel := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "users", "12", "roles"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "users", "12", "roles"),
		},
	}
	rel2 := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "users", "21", "roles"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "users", "21", "roles"),
		},
	}
	exstruct := &jsapi.Document{
		Links: &jsapi.Links{Self: fmt.Sprintf("%s/%s", bslink, "users")},
		Data: &jsapi.DataContainer{
			DataArray: []jsapi.Data{
				jsapi.Data{
					Type:          "users",
					ID:            "12",
					Attributes:    []byte(`{"name":"Caboose","email":"caboose@caboose.com"}`),
					Relationships: map[string]jsapi.Relationship{"roles": rel},
					Links:         &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s", bslink, "users", "12")},
				},
				jsapi.Data{
					Type:          "users",
					ID:            "21",
					Attributes:    []byte(`{"name":"Damon","email":"damon@damon.com"}`),
					Relationships: map[string]jsapi.Relationship{"roles": rel2},
					Links:         &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s", bslink, "users", "21")},
				},
			},
		},
	}
	if !reflect.DeepEqual(pstruct, exstruct) {
		t.Fatal("expected and generated jsonapi structure did not match")
	}
}

func TestPaginationLinks(t *testing.T) {
	srvinfo := aphtest.NewTestApiInfo()
	bslink := generateBaseLink(srvinfo)
	users := []*User{
		&User{
			ID:    "12",
			Name:  "Caboose",
			Email: "caboose@caboose.com",
		},
		&User{
			ID:    "21",
			Name:  "Damon",
			Email: "damon@damon.com",
		},
	}

	pageOpt := &pagination.Props{
		Records: 100,
		Entries: 10,
		Current: 5,
	}
	pstruct, err := MarshalWithPagination(users, srvinfo, pageOpt)
	if err != nil {
		t.Errorf("error in marshaling to structure %s\n", err)
	}
	rel := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "users", "12", "roles"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "users", "12", "roles"),
		},
	}
	rel2 := jsapi.Relationship{
		Links: &jsapi.Links{
			Self:    fmt.Sprintf("%s/%s/%s/relationships/%s", bslink, "users", "21", "roles"),
			Related: fmt.Sprintf("%s/%s/%s/%s", bslink, "users", "21", "roles"),
		},
	}
	resourceLink := fmt.Sprintf("%s/%s", bslink, "users")
	pageLinks := &jsapi.Links{
		Self:     fmt.Sprintf("%s?page[number]=%d&page[size]=%d", resourceLink, 5, 10),
		First:    fmt.Sprintf("%s?page[number]=%d&page[size]=%d", resourceLink, 1, 10),
		Previous: fmt.Sprintf("%s?page[number]=%d&page[size]=%d", resourceLink, 4, 10),
		Last:     fmt.Sprintf("%s?page[number]=%d&page[size]=%d", resourceLink, 10, 10),
		Next:     fmt.Sprintf("%s?page[number]=%d&page[size]=%d", resourceLink, 6, 10),
	}
	exstruct := &jsapi.Document{
		Links: pageLinks,
		Data: &jsapi.DataContainer{
			DataArray: []jsapi.Data{
				jsapi.Data{
					Type:          "users",
					ID:            "12",
					Attributes:    []byte(`{"name":"Caboose","email":"caboose@caboose.com"}`),
					Relationships: map[string]jsapi.Relationship{"roles": rel},
					Links:         &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s", bslink, "users", "12")},
				},
				jsapi.Data{
					Type:          "users",
					ID:            "21",
					Attributes:    []byte(`{"name":"Damon","email":"damon@damon.com"}`),
					Relationships: map[string]jsapi.Relationship{"roles": rel2},
					Links:         &jsapi.Links{Self: fmt.Sprintf("%s/%s/%s", bslink, "users", "21")},
				},
			},
		},
		Meta: map[string]interface{}{
			"pagination": map[string]int{
				"records": 100,
				"total":   10,
				"size":    10,
				"number":  5,
			},
		},
	}
	if !reflect.DeepEqual(pstruct, exstruct) {
		t.Fatal("expected and generated jsonapi structure did not match")
	}
}

func TestIncludes(t *testing.T) {
	srvinfo := aphtest.NewTestApiInfo()
	bslink := generateBaseLink(srvinfo)
	u := &User{
		ID:    "32",
		Name:  "Tucker",
		Email: "tucker@jumbo.com",
		Roles: []*Role{
			&Role{
				ID:          "44",
				Role:        "Administrator",
				Description: "The God",
			},
			&Role{
				ID:          "45",
				Role:        "Management",
				Description: "The collector",
			},
		},
	}
	pstruct, err := MarshalToStructWrapper(u, srvinfo)
	if err != nil {
		t.Fatalf("error in marshaling to structure %s\n", err)
	}
	if len(pstruct.Included) != 2 {
		b, _ := json.Marshal(pstruct)
		t.Fatalf("no included members %s\n", string(b))
	}
	istruct := pstruct.Included[0]
	if istruct.Type != "roles" {
		t.Errorf("actual %s expected %s type\n", istruct.Type, "roles")
	}
	explink := istruct.Relationships["users"].Links.Related
	rlink := fmt.Sprintf("%s/%s/%s/%s", bslink, "roles", istruct.ID, "consumers")
	if explink != rlink {
		t.Errorf("actual %s link expected %s", rlink, explink)
	}
}

func TestGetTypeName(t *testing.T) {
	r := &Role{
		ID:          "44",
		Role:        "Administrator",
		Description: "The God",
	}
	u := &User{
		ID:    "32",
		Name:  "Tucker",
		Email: "tucker@jumbo.com",
	}
	rtype := GetTypeName(r)
	if rtype != "roles" {
		t.Errorf("actual type %s did not match the expect % type name", rtype, "roles")
	}
	utype := GetTypeName(u)
	if utype != "users" {
		t.Errorf("actual type %s did not match the expect % type name", utype, "users")
	}
}

func TestAttributeFields(t *testing.T) {
	u := &User{
		ID:    "32",
		Name:  "Tucker",
		Email: "tucker@jumbo.com",
	}
	attrs := GetAttributeFields(u)
	if len(attrs) != 2 {
		t.Fatalf("actual no of attributes %d does not match the expected %d length", len(attrs), 2)
	}
	for _, n := range []string{"name", "email"} {
		if !aphcollection.Contains(attrs, n) {
			t.Errorf("attribute %s could not be found in collection %s", n, strings.Join(attrs, "|"))
		}
	}
}

func TestAllRelationships(t *testing.T) {
	u := &User{
		ID:    "32",
		Name:  "Tucker",
		Email: "tucker@jumbo.com",
	}
	rels := GetAllRelationships(u)
	if len(rels) != 2 {
		t.Fatalf("actual no of relationships %d does not match the expected %d relationships", len(rels), 2)
	}
	for _, r := range rels {
		if r.Name != "roles" {
			t.Errorf("actual name %s does not match the expected %s one", r.Name, "roles")
		}
	}
}
