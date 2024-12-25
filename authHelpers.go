package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword takes a password string and returns a hash
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// checkPasswordHash compares a password to a hash and returns true if they
// match
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// newClaims creates a new set of claims using user credentials, and uses
// the claims to create a new token using renewToken()
func newClaims(w http.ResponseWriter, r *http.Request, c *credentials) (ctxx context.Context) {
	claims := credentials{
		c.Name,
		"",
		true,
		[]string{},
		0,
		jwt.StandardClaims{
			// ExpiresAt: 15000,
			// Issuer:    "test",
		},
	}

	return renewToken(w, r, &claims)
}

// parseToken takes a token string, checks its validity, and parses it into a
// set of credentials. If the token is invalid it returns an error
func parseToken(tokenString string) (*credentials, error) {
	var claims *credentials
	token, err := jwt.ParseWithClaims(tokenString, &credentials{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if err != nil {
		log.Println(err)
		cc := credentials{IsLoggedIn: false}
		return &cc, err
	}
	if claims, ok := token.Claims.(*credentials); ok && token.Valid {
		return claims, nil
	}
	return claims, err
}

// renewToken renews a users token using existing claims, sets it as a cookie
// on the client, and adds it to the database.
// TODO: FIX EXPIRY
func renewToken(w http.ResponseWriter, r *http.Request, claims *credentials) (ctxx context.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		log.Println(err)
	}

	expire := time.Now().Add(10 * time.Minute)
	cookie := http.Cookie{Name: "token", Value: ss, Path: "/", Expires: expire, MaxAge: 0}
	http.SetCookie(w, &cookie)

	rdb.Set(rdx, claims.Name+":token", ss, 0)
	ctxx = context.WithValue(r.Context(), ctxkey, claims)
	return
}

// marshalCredentials is used convert a request body into a credentials{}
// struct
func marshalCredentials(r *http.Request) (*credentials, error) {
	t := &credentials{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(t)
	if err != nil {
		return t, err
	}
	return t, nil
}
