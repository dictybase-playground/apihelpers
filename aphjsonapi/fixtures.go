package aphjsonapi

import "fmt"
import "github.com/manyminds/api2go/jsonapi"

type Permission struct {
	ID          string `json:"-"`
	Permission  string `json:"permission"`
	Description string `json:"description"`
}

func (p Permission) GetID() string {
	return p.ID
}

type Role struct {
	ID          string        `json:"-"`
	Role        string        `json:"role"`
	Description string        `json:"description"`
	Permissions []*Permission `json:"-"`
	Users       []*User       `json:"-"`
}

func (r *Role) GetID() string {
	return r.ID
}

func (r *Role) GetSelfLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{Name: "users", Type: "users"},
		RelationShipLink{Name: "permissions", Type: "permissions"},
	}
}

func (r *Role) ValidateSelfLinks() error {
	return nil
}

func (r *Role) GetRelatedLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{
			Name:           "users",
			SuffixFragment: fmt.Sprintf("%s/%s/%s", "roles", r.GetID(), "consumers"),
			Type:           "users",
		},
		RelationShipLink{Name: "permissions", Type: "users"},
	}
}

func (r *Role) ValidateRelatedLinks() error {
	return nil
}

type User struct {
	ID    string  `json:"-"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Roles []*Role `json:"-"`
}

// GetID satisfies jsonapi.MarshalIdentifier interface
func (u *User) GetID() string {
	return u.ID
}

// GetSelfLinksInfo satisfies MarshalSelfRelations interface
func (u *User) GetSelfLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{Name: "roles", Type: "roles"},
	}
}

func (u *User) ValidateSelfLinks() error {
	return nil
}

// GetSelfLinksInfo satisfies MarshalRelatedRelations interface
func (u *User) GetRelatedLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{Name: "roles", Type: "roles"},
	}
}

func (u *User) ValidateRelatedLinks() error {
	return nil
}

// GetReferences satisfies jsonapi.MarshalReferences interface
func (u *User) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		jsonapi.Reference{Type: "roles", Name: "roles"},
	}
}

// GetReferencedStructs satisfies jsonapi.MarshalIncludedRelations interface
func (u *User) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	var result []jsonapi.MarshalIdentifier
	for _, r := range u.Roles {
		result = append(result, r)
	}
	return result
}
