package models

type UpworkAccount struct {
	Email    string
	Password string
	TwoFA    string
	Token    string
	Cookie   []Cookie
}
