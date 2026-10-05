package config

// this is a example file for the gitignored secrets.go file containing secrets.
// variables are unexported by intention to not have collisions with the real file

const resendApiKey = "re_....." // api key for resend email service

const jwtHMACKey = "...." // bytes

const _dbUser = "..."
const _dbPassword = "..."
const _dbHost = "localhost"
const _dbPort = "5432"
const _dbName = "..."
const _dbSslMode = "disable"

// database connection string
const dbConnStr = "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=" + dbSslMode
