package aphrender

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dictyBase/apihelpers/aphtest"
	"github.com/dictyBase/go-middlewares/middlewares/pagination"
)

type Permission struct {
	ID          string `json:"-"`
	Permission  string `json:"permission"`
	Description string `json:"description"`
}

func (p Permission) GetID() string {
	return p.ID
}

func TestResource(t *testing.T) {
	jsonBlob := []byte(`
		{
			"links": {
				"self":"https://api.dictybase.org/1.0/permissions/10"
			},
			"data": {
				"type": "permissions",
				"id": "10",
				"attributes": {
					"permission": "gene curation",
					"description": "Authority to edit gene information"
				}
			}
		}
	`)
	perm := &Permission{
		ID:          "10",
		Permission:  "gene curation",
		Description: "Authority to edit gene information",
	}
	srvinfo := aphtest.NewTestApiInfo()
	w := httptest.NewRecorder()
	Resource(perm, srvinfo, w)

	if w.Header().Get("Content-Type") != "application/vnd.api+json" {
		t.Fatalf("no jsonapi response header %s", w.Header().Get("Content-Type"))
	}
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code %d\n", w.Code)
	}
	expJson := aphtest.IndentJSON(jsonBlob)
	matchJson := aphtest.IndentJSON(w.Body.Bytes())
	if bytes.Compare(expJson, matchJson) != 0 {
		t.Fatalf("expected \n%s jsonapi error response does not match with \n%s\n", string(expJson), string(matchJson))
	}
}

func TestResourceCollection(t *testing.T) {
	jsonBlob := []byte(`
		{
			"links": {
				"self":"https://api.dictybase.org/1.0/permissions?page[number]=5&page[size]=10",
				"first":"https://api.dictybase.org/1.0/permissions?page[number]=1&page[size]=10",
				"prev":"https://api.dictybase.org/1.0/permissions?page[number]=4&page[size]=10",
				"next":"https://api.dictybase.org/1.0/permissions?page[number]=6&page[size]=10",
				"last":"https://api.dictybase.org/1.0/permissions?page[number]=10&page[size]=10"
			},
			"data": [{
				"type": "permissions",
				"id": "10",
				"attributes": {
					"permission": "gene curation",
					"description": "Authority to edit gene information"
				},
				"links": {
					"self": "https://api.dictybase.org/1.0/permissions/10"
				}
			}],
			"meta": {
				"pagination": {
					"number": 5,
					"records": 100,
					"size": 10,
					"total": 10
				}
			}
		}
	`)
	pageOpt := &pagination.Props{
		Records: 100,
		Entries: 10,
		Current: 5,
	}
	srvinfo := aphtest.NewTestApiInfo()
	permissions := []*Permission{
		&Permission{
			ID:          "10",
			Permission:  "gene curation",
			Description: "Authority to edit gene information",
		},
	}
	w := httptest.NewRecorder()
	ResourceCollection(permissions, srvinfo, w, pageOpt)
	if w.Header().Get("Content-Type") != "application/vnd.api+json" {
		t.Fatalf("no jsonapi response header %s", w.Header().Get("Content-Type"))
	}
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code %d\n", w.Code)
	}
	expJson := aphtest.IndentJSON(jsonBlob)
	matchJson := aphtest.IndentJSON(w.Body.Bytes())
	if bytes.Compare(expJson, matchJson) != 0 {
		t.Fatalf("expected \n%s jsonapi error response does not match with \n%s\n", string(expJson), string(matchJson))
	}
}
