package entity

import (
	"time"
	"uuid"
)

// used to describe account status
type AccountStatus string

const (
	StatusUnverified AccountStatus = "unverified"
	StatusVerified   AccountStatus = "verified"
	StatusBanned     AccountStatus = "banned"
)

// AuthChannel defines WHAT medium is being used.
type AuthChannel string

const (
	ChannelEmail    AuthChannel = "email"
	ChannelPhone    AuthChannel = "phone"
	ChannelUsername AuthChannel = "username" // Used for DB lookup only/ login
	ChannelTOTP     AuthChannel = "totp"     // Used for 2FA validation only
)

type User struct {
	// user data
	ID           uuid.UUID     `db:"id" json:"id" db_constraints:"primary key default gen_random_uuid()"`
	Email        *string       `db:"email" json:"email" db_constraints:"unique"`
	Phone        *string       `db:"phone" json:"phone" db_constraints:"unique"`
	PasswordHash string        `db:"password_hash" json:"password_hash" db_constraints:"not null"`
	Dob          string        `db:"dob" json:"dob" db_constraints:"not null"`
	Status       AccountStatus `db:"status" json:"status" db_constraints:"not null default 'unverified'"`
	// 2fa related
	TotpSecretKey *string       `db:"totp_secret_key" json:"-"`
	TwoFAs        []AuthChannel `db:"two_fas" json:"-"` // slice of auth channels, if nil means twoFa not enabled, else its enabled on whichever identifiers are in the slice
	// timestamps
	CreatedAt time.Time `db:"created_at" json:"-" db_constraints:"not null default current_timestamp"`
	UpdatedAt time.Time `db:"updated_at" json:"-" db_constraints:"not null default current_timestamp"`
}

type UserSession struct {
	// session info
	ID        uuid.UUID `db:"id" db_constraints:"primary key default gen_random_uuid()"`
	JwtID     uuid.UUID `db:"jwt_id" db_constraints:"not null unique"`
	UserID    uuid.UUID `db:"user_id" db_constraints:"not null references users(id)"`
	TokenHash string    `db:"token_hash"  db_constraints:"not null"`
	// device info
	IPAddress   *string `db:"ip_address"`
	OSName      *string `db:"os_name"`
	BrowserName *string `db:"browser_name"`
	DeviceType  *string `db:"device_type"`
	// timestamps
	ExpiresAt time.Time `db:"expires_at" db_constraints:"not null"`
	CreatedAt time.Time `db:"created_at" db_constraints:"not null default current_timestamp"`
}

type Profile struct {
	// profile info
	UserID      uuid.UUID `db:"user_id" json:"user_id" db_constraints:"primary key references users(id)"`
	Username    string    `db:"username" json:"username" db_constraints:"not null unique"`
	DisplayName *string   `db:"display_name" json:"display_name"`
	AvatarUrl   *string   `db:"avatar_url" json:"avatar_url"`
	IsPrivate   bool      `db:"is_private" json:"is_private" db_constraints:"not null default false"`
	Bio         *string   `db:"bio" json:"bio"`
	// timestamps
	UpdatedAt time.Time `db:"updated_at" json:"-" db_constraints:"not null default current_timestamp"`
}
