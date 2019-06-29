package main

import (
	"crypto/rsa"
	"errors"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

func parseJWT(t string, keys *jwk.Set) (*jwt.Token, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		keys := keys.LookupKeyID(token.Header["kid"].(string))
		if len(keys) == 0 {
			log.Println("failed to look up JWKs")
			return nil, errors.New("could not find matching `kid` in well known tokens")
		}
		// Build the public RSA key
		key, err := keys[0].Materialize()
		if err != nil {
			log.Printf("failed to create public key: %s", err)
			return nil, err
		}
		rsaPublicKey := key.(*rsa.PublicKey)
		return rsaPublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func validateJWT(token *jwt.Token, aud string) (*jwt.Token, error) {
	if !token.Valid {
		err := errors.New("invalid token")
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Then check time based claims; exp, iat, nbf
		err := claims.Valid()
		if err == nil {
			// Then check that `aud` matches the app client id
			// (if `aud` even exists on the token, second arg is a "required" option)
			if claims.VerifyAudience(aud, false) {
				return token, nil
			}

			err = errors.New("token audience does not match client id")
			return nil, err
		}

		return nil, err
	}

	err := errors.New("issue parsing claims in token")
	return nil, err
}
