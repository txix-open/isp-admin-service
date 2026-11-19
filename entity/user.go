package entity

import "time"

type User struct {
	OauthUserId          *string
	Id                   int64
	Roles                []int
	FirstName            string
	LastName             string
	FullName             string
	Description          string
	Email                string
	Password             string
	Blocked              bool
	LastActiveAt         time.Time
	UpdatedAt            time.Time
	CreatedAt            time.Time
	LastSessionCreatedAt *time.Time
}

type OauthUser struct {
	RoleIds     []int
	OauthUserId string
	FirstName   string
	LastName    string
	Email       string
	Description string
	FullName    string
}

type UpdateUser struct {
	FirstName   string
	LastName    string
	FullName    string
	Email       string
	Description string
}
