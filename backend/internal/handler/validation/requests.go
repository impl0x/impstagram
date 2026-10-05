package validations

type strValue struct {
	Value string `json:"value" validation:"required"`
}
