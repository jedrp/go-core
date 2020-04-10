package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}
type JwtValidator struct {
	jwks        *Jwks
	Aud         string
	Issuer      string
	JwkAddress  string
	cert        string
	certRenewAt time.Time
	mux         sync.Mutex
}

func NewJwtValidator(aud string, issuer string) (*JwtValidator, error) {
	v := &JwtValidator{
		Aud:        aud,
		Issuer:     issuer,
		JwkAddress: fmt.Sprintf("%s/.well-known/openid-configuration/jwks", issuer),
	}
	jwks, e := v.GetJwks()
	v.jwks = jwks
	return v, e
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func (config *JwtValidator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, config.ValidationKeyGetter)
}

func (config *JwtValidator) ValidationKeyGetter(token *jwt.Token) (interface{}, error) {
	audClaim, _ := token.Claims.(jwt.MapClaims)["aud"]
	validAud := false
	if real, ok := audClaim.([]interface{}); ok {
		for _, v := range real {
			if v == config.Aud {
				validAud = true
			}
		}
	} else {
		if v, ok := audClaim.(interface{}); ok {
			if v == config.Aud {
				validAud = true
			}
		}
	}
	if !validAud {
		return token, errors.New("Invalid audience.")
	}
	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(config.Issuer, false)
	if !checkIss {
		return token, errors.New("Invalid issuer.")
	}

	cert, err := config.getPemCert(token)
	if err != nil {
		panic(err.Error())
	}

	result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	return result, nil
}

func (config *JwtValidator) getPemCert(token *jwt.Token) (string, error) {
	if config.cert != "" && !config.certRenewAt.IsZero() && time.Now().Before(config.certRenewAt) {
		return config.cert, nil
	}
	config.mux.Lock()
	defer config.mux.Unlock()
	if config.cert != "" && !config.certRenewAt.IsZero() && time.Now().Before(config.certRenewAt) {
		return config.cert, nil
	}

	for _, v := range config.jwks.Keys {
		if token.Header["kid"] == v.Kid {
			config.cert = "-----BEGIN CERTIFICATE-----\n" + v.X5c[0] + "\n-----END CERTIFICATE-----"
		}
		config.certRenewAt = time.Now().Add(15 * time.Minute)
	}

	if config.cert == "" {
		err := errors.New("unable to find appropriate key")
		return config.cert, err
	}
	return config.cert, nil
}

func (config *JwtValidator) GetJwks() (*Jwks, error) {
	resp, err := http.Get(config.JwkAddress)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jwks = &Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return nil, err
	}
	return jwks, nil
}
