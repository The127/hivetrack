package models

import (
	"time"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
)

type UserChange int

const (
	UserChangeSub         UserChange = iota
	UserChangeEmail       UserChange = iota
	UserChangeDisplayName UserChange = iota
	UserChangeAvatarURL   UserChange = iota
	UserChangeIsAdmin     UserChange = iota
	UserChangeLastLoginAt UserChange = iota
)

type User struct {
	BaseModel
	change.List[UserChange]

	sub         string
	email       string
	displayName string
	avatarURL   *string
	isAdmin     bool
	lastLoginAt time.Time
}

func NewUser(sub, email, displayName string) *User {
	return &User{
		BaseModel:   NewBaseModel(),
		List:        change.NewList[UserChange](),
		sub:         sub,
		email:       email,
		displayName: displayName,
	}
}

func NewUserFromDB(id uuid.UUID, createdAt time.Time, version any,
	sub, email, displayName string, avatarURL *string, isAdmin bool, lastLoginAt time.Time) *User {
	return &User{
		BaseModel:   NewBaseModelFromDB(id, createdAt, createdAt, version),
		List:        change.NewList[UserChange](),
		sub:         sub,
		email:       email,
		displayName: displayName,
		avatarURL:   avatarURL,
		isAdmin:     isAdmin,
		lastLoginAt: lastLoginAt,
	}
}

func (u *User) GetSub() string            { return u.sub }
func (u *User) GetEmail() string          { return u.email }
func (u *User) GetDisplayName() string    { return u.displayName }
func (u *User) GetAvatarURL() *string     { return u.avatarURL }
func (u *User) GetIsAdmin() bool          { return u.isAdmin }
func (u *User) GetLastLoginAt() time.Time { return u.lastLoginAt }

func (u *User) SetSub(v string) {
	if u.sub == v {
		return
	}
	u.sub = v
	u.TrackChange(UserChangeSub)
}

func (u *User) SetEmail(v string) {
	if u.email == v {
		return
	}
	u.email = v
	u.TrackChange(UserChangeEmail)
}

func (u *User) SetDisplayName(v string) {
	if u.displayName == v {
		return
	}
	u.displayName = v
	u.TrackChange(UserChangeDisplayName)
}

func (u *User) SetAvatarURL(v *string) {
	u.avatarURL = v
	u.TrackChange(UserChangeAvatarURL)
}

func (u *User) SetIsAdmin(v bool) {
	if u.isAdmin == v {
		return
	}
	u.isAdmin = v
	u.TrackChange(UserChangeIsAdmin)
}

func (u *User) SetLastLoginAt(v time.Time) {
	u.lastLoginAt = v
	u.TrackChange(UserChangeLastLoginAt)
}
