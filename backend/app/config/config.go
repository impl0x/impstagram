package config

// contains constant values that should not be injected as environment variables

type Database uint8

const (
	Postgres Database = iota
)

const CurrentDatabase = Postgres

type PasswordHasher uint8

const (
	Argon2ID PasswordHasher = iota
)

const CurrentHasher = Argon2ID

type Transport uint8

const (
	HTTP Transport = iota
)

const CurrentTransport = HTTP

type EmailSender uint8

const (
	Resend EmailSender = iota
)

const CurrentEmailSender = Resend
