package models

type FindUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}
