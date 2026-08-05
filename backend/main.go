package main

import (
	"backend/internal/features/auth"
	"encoding/json"

	"github.com/impl0x/mo/validator"
)

func main() {
	demoSigninReq:=auth.SigninRequest{
		Username: "dd",
		Password: "test",
	}
	err:=validator.Validate(demoSigninReq)
	if err!=nil{
		for _,e :=range err.Errors{
			b,_:=json.MarshalIndent(e.JsonFormat(),"","    ")
			println(string(b))
		}
	}

}

