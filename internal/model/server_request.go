package model

type ServerRequest struct {
	Name         string `json:"name,omitempty"`
	UserId       string `json:"userId"`
	Slug         string `json:"slug"`
	OwnedSeconds int    `json:"ownedSeconds"`
}
