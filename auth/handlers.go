package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"os"
	"time"
)

var Dbpool *pgxpool.Pool // todo: solve the problem with concurrent access
var jwtKey = []byte("https://open.spotify.com/playlist/3YFOAjcM51r1WUnbWtxWh0?si=92140f94c6")

type Register struct {
	Username string `json:"username"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	Username  string
	Email string
	Password string
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenData struct {
	AccessToken string `json:"token"`
	ExpirationTime time.Time `json:"expiration"`
}

type ClaimData struct {
	AccessToken string `json:"token"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func getClaims(cntx *gin.Context) (*Claims, *jwt.Token, int) {
	var claimData ClaimData
	err := json.NewDecoder(cntx.Request.Body).Decode(&claimData)
	if err != nil {
		return nil, nil, http.StatusBadRequest
	}
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(claimData.AccessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, nil, http.StatusUnauthorized
		}
		return nil, nil, http.StatusBadRequest
	}
	if !tkn.Valid {
		return nil, nil, http.StatusUnauthorized
	}
	if claims.ExpiresAt.Time.Sub(time.Now()) > 30 * time.Minute {
		return nil, nil, http.StatusUnauthorized
	}
	return claims, tkn, http.StatusOK
}

// Signup godoc
// @Summary Signup Summary
// @Schemes
// @Description Signup Description
// @Tags auth
// @Accept json
// @Produce json
// @Param id body Register true "Register"
// @Success 200
// @Router /auth/signup [post]
func Signup(cntx *gin.Context) {

	var creds Register
	err := json.NewDecoder(cntx.Request.Body).Decode(&creds)
	if err != nil {
		cntx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := Dbpool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	user_ulid := fmt.Sprintf("%s", CreateULID())
	user_secret, err := HashPassword(creds.Password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error create password hash:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	row := conn.QueryRow(context.Background(),
		"insert into auth.user (user_ulid, user_name, user_secret, email) values ($1, $2, $3, $4) returning user_ulid",
		user_ulid, creds.Username, user_secret, creds.Email)
	var id string
	err = row.Scan(&id)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to insert:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// todo: delete after debugging
	cntx.Writer.Write([]byte(id))
}

// Delete godoc
// @Summary Delete Summary
// @Schemes
// @Description Delete Description
// @Tags auth
// @Accept json
// @Produce json
// @Param id body ClaimData true "Credentials"
// @Success 200
// @Router /auth/delete [post]
func Delete(cntx *gin.Context) {

	claims, _, status := getClaims(cntx)
	if status != http.StatusOK {
		cntx.Writer.WriteHeader(status)
		return
	}

	conn, err := Dbpool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"delete from auth.user where user_name=$1 returning user_ulid", claims.Username)
	var id string
	err = row.Scan(&id)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to delete:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// todo: delete after debugging
	cntx.Writer.Write([]byte(id))
}

// Signin godoc
// @Summary Signin Summary
// @Schemes
// @Description Signin Description
// @Tags auth
// @Accept json
// @Produce json
// @Param id body Credentials true "Credentials"
// @Success 200 {object} TokenData
// @Router /auth/signin [post]
func Signin(cntx *gin.Context) {

	var creds Credentials
	err := json.NewDecoder(cntx.Request.Body).Decode(&creds)
	if err != nil {
		cntx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := Dbpool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error acquiring connection:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	var users []*User
	err = pgxscan.Select(context.Background(), conn, &users,
		"select user_name as username, email, user_secret as password from auth.user where user_name = $1",
		creds.Username)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error select user info:", err)
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		cntx.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	secret := users[0].Password
	if !CheckPasswordHash(creds.Password, secret) {
		cntx.Writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(jwtKey)
	if err != nil {
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenData := TokenData{accessToken, expirationTime}
	// todo: refactoring to Marshaling
	cntx.JSON(http.StatusOK, gin.H{"token": tokenData.AccessToken, "expiration": tokenData.ExpirationTime})
}

// Welcome godoc
// @Summary Welcome Summary
// @Schemes
// @Description Welcome Description
// @Tags auth
// @Accept json
// @Produce json
// @Param id body ClaimData true "Credentials"
// @Success 200 {string} Welcome
// @Router /auth/welcome [get]
func Welcome(cntx *gin.Context) {

	claims, _, status := getClaims(cntx)
	if status != http.StatusOK {
		cntx.Writer.WriteHeader(status)
		return
	}

	cntx.Writer.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

// Refresh godoc
// @Summary Refresh Summary
// @Schemes
// @Description Refresh Description
// @Tags auth
// @Accept json
// @Produce json
// @Param id body Credentials true "Credentials"
// @Success 200 {object} TokenData
// @Router /auth/refresh [post]
func Refresh(cntx *gin.Context) {

	claims, _, status := getClaims(cntx)
	if status != http.StatusOK {
		cntx.Writer.WriteHeader(status)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(jwtKey)
	if err != nil {
		cntx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenData := TokenData{accessToken, expirationTime}
	// todo: refactoring to Marshaling
	cntx.JSON(http.StatusOK, gin.H{"token": tokenData.AccessToken, "expiration": tokenData.ExpirationTime})
}
