package config

import "backend/pkg/cryptoutil"

// contains constant values that should not be injected as environment variables

// App version according to release
const AppVersion = "0.5.0"

type current struct {
	Database  Database
	Transport Transport
	Extra     extra
	Hasher    hasher
}

type hasher struct {
	Password     HashingAlgorithm
	RefreshToken HashingAlgorithm
}

type extra struct {
	EmailSender      EmailSender
	RequestIDBitSize cryptoutil.BitSize // Bit size for request IDs
}

// Current values for the app configurations
var _current = current{
	Database:  Postgres,
	Transport: HTTP,
	Extra: extra{
		EmailSender:      Resend,
		RequestIDBitSize: cryptoutil.Bit64,
	},
	Hasher: hasher{
		Password:     Argon2ID,
		RefreshToken: MD5,
	},
}

// Current app configuration
var Current = func() current { return _current } // function returning a copy of [_current], preventing it from being modified
