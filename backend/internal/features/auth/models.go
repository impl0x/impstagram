package auth

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=15"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`
	Dob      string `json:"dob" validate:"required"`

	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"optional,min=3,max=15"`
	Email    string `json:"email" validate:"optional,email"`
	Phone    string `json:"phone" validate:"optional,e.164"`

	Password string `json:"password" validate:"required,min=8"`
}
