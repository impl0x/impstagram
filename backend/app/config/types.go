package config

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

// Hashing algorithms used to hash passwords,tokens,etc in the database
type HashingAlgorithm uint8

const (
	Argon2ID HashingAlgorithm = iota
	MD5
	SHA256
)

// Implementation used to send emails to users
type EmailSender uint8

const (
	Resend EmailSender = iota
)
