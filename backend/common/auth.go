package common

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// https://pkg.go.dev/github.com/golang-jwt/jwt/v5#example-Parse-Hmac
// https://github.com/dgrijalva/jwt-go/blob/master/hmac_example_test.go

func (cfg *SrvCfg) ValidateJwt(jwtToken string) bool {
	token, err := jwt.Parse(jwtToken, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("Couldn't sign using %v", t.Header["alg"])
		}

		return cfg.JWTSecret, nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse your jwt with error %v\n", err)
		return false
	}

	_, ok := token.Claims.(jwt.MapClaims)

	// I presume this checks if the token is valid?
	if ok && token.Valid {
		return true
	} else {
		return false
	}
}
