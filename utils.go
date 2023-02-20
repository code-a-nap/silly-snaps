package main

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey = []byte("superSecretJWTKey!")

type JWTClaim struct {
	Email  string `json:"email"`
	Role   string `json:"role"`
	Secret string `json:"secret"`
	jwt.StandardClaims
}

func openLogFile(logfile string) {
	if logfile != "" {
		lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

		if err != nil {
			log.Fatal("OpenLogfile: os.OpenFile:", err)
		}
		log.SetOutput(lf)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func GenerateJWT(email string, role string, secret string) (tokenString string, err error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	claims := &JWTClaim{
		Email:  email,
		Role:   role,
		Secret: secret,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(secretKey)
	return tokenString, err
}

func ValidateToken(signedToken string) (c *JWTClaim, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return nil, err
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return nil, err
	}
	return claims, nil
}

func checkEnvVars() error {
	if os.Getenv("JWT_KEY") != "" {
		secretKey = []byte(os.Getenv("JWT_KEY"))
	}
	if os.Getenv("USER_PWD") == "" || os.Getenv("ADMIN_PWD") == "" || os.Getenv("FLAG") == "" {
		return errors.New("environment variables missing")
	}
	return nil
}
