package jwt

// todo
type Jwt struct {
	key string
}

func (j Jwt) GenerateToken() string {
	return "token"
}
