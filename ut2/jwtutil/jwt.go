package jwtutil

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	P int64
	E int64
	S int64
	D string `json:",omitempty"`
}

func (m *Claims) GetSign() int64 {
	return (m.E>>8)%(m.P/10000) + 996
}

func (m *Claims) Valid() error {

	exp := time.Unix(m.E, 0)
	if time.Now().After(exp) {
		return jwt.ErrTokenExpired
	}

	s := m.GetSign()
	if s != m.S {
		return jwt.ErrTokenSignatureInvalid
	}

	return nil
}

var signatureKey = []byte("super_无敌_pwd##A")

func SetKey(key string) {
	signatureKey = []byte(key)
}

func NewToken(pid int64, expAt time.Time) (string, error) {
	return NewTokenWithData(pid, expAt, "")
}

func NewTokenWithData(pid int64, expAt time.Time, data string) (string, error) {

	claims := &Claims{
		P: pid,
		E: expAt.Unix(),
		D: data,
	}

	claims.S = claims.GetSign()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(signatureKey)
	if err != nil {
		return "", err
	}
	idx := strings.IndexByte(s, '.')
	return s[idx+1:], nil
}

func ParseClaims(tokenString string) (*Claims, error) {
	tokenString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." + tokenString
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return signatureKey, nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func ParseTokenData(tokenString string) (int64, string, error) {
	claims, err := ParseClaims(tokenString)
	if err != nil {
		return 0, "", err
	}

	return claims.P, claims.D, nil
}

func ParseToken(tokenString string) (int64, error) {
	pid, _, err := ParseTokenData(tokenString)
	return pid, err
}
