package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserClaims adalah payload yang akan dimasukkan ke dalam JWT Access Token
type UserClaims struct {
	UserID       string `json:"user_id"`
	NIK          string `json:"nik"`
	Role         string `json:"role"`
	DepartmentID int    `json:"department_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken membuat JWT berumur pendek (15 Menit) untuk Frontend
func GenerateAccessToken(userID, nik, role string, deptID int, secretKey string) (string, error) {
	claims := UserClaims{
		UserID:       userID,
		NIK:          nik,
		Role:         role,
		DepartmentID: deptID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 Menit
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GenerateRefreshToken membuat string random unik (opaque token) untuk cookie & session DB
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
