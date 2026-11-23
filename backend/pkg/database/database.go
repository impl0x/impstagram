package database

// Represents a database connection manager
type Database interface{
	// Closes the database connection
	Close()
}