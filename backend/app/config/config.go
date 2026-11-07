package config

// contains constant values that should not be injected as environment variables

// App version according to release
const AppVersion = "0.5.0"

// database type represents a database storage which implements repository
type Database uint8

const (
	// Postgres sql
	Postgres Database = iota
)

// Transport refers to the protocol used to talk to the frontend as a server
type Transport uint8

const (
	HTTP Transport = iota
)

// Hashing algorithm used to hash passwords in the database
type PasswordHasher uint8

const (
	Argon2ID PasswordHasher = iota
)

// Implementation used to send emails to users
type EmailSender uint8

const (
	Resend EmailSender = iota
)

// Current values for the app configurations
const (
	CurrentDatabase       = Postgres
	CurrentTransport      = HTTP
	CurrentPasswordHasher = Argon2ID
	CurrentEmailSender    = Resend
)
