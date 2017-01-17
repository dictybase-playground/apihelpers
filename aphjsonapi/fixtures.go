package aphjsonapi

import "fmt"

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
		RelationShipLink{Name: "users"},
		RelationShipLink{Name: "permissions"},
	}
}

func (r *Role) GetRelatedLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{
			Name:           "users",
			SuffixFragment: fmt.Sprintf("%s/%s/%s", "roles", r.GetID(), "consumers"),
		},
		RelationShipLink{Name: "permissions"},
	}
}

type User struct {
	ID    string  `json:"-"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Roles []*Role `json:"-"`
}

func (u *User) GetID() string {
	return u.ID
}

func (u *User) GetSelfLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{Name: "roles"},
	}
}

func (u *User) GetRelatedLinksInfo() []RelationShipLink {
	return []RelationShipLink{
		RelationShipLink{Name: "roles"},
	}
}
