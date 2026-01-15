package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dowglassantana/product-redis-api/internal/infrastructure/config"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type authContextKey string

const UserContextKey authContextKey = "user"

type UserClaims struct {
	Subject           string   `json:"sub"`
	Email             string   `json:"email"`
	PreferredUsername string   `json:"preferred_username"`
	RealmRoles        []string `json:"realm_roles"`
}

type JWTAuth struct {
	keycloakConfig *config.KeycloakConfig
	logger         *zap.Logger
	jwks           *JWKS
	jwksMutex      sync.RWMutex
	lastFetch      time.Time
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewJWTAuth(keycloakConfig *config.KeycloakConfig, logger *zap.Logger) *JWTAuth {
	return &JWTAuth{
		keycloakConfig: keycloakConfig,
		logger:         logger,
	}
}

func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			j.unauthorizedResponse(w, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			j.unauthorizedResponse(w, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		claims, err := j.validateToken(tokenString)
		if err != nil {
			j.logger.Debug("token validation failed", zap.Error(err))
			j.unauthorizedResponse(w, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (j *JWTAuth) validateToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		return j.getPublicKey(kid)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Validate issuer
	iss, _ := mapClaims["iss"].(string)
	if iss != j.keycloakConfig.Issuer() {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", j.keycloakConfig.Issuer(), iss)
	}

	userClaims := &UserClaims{
		Subject:           getString(mapClaims, "sub"),
		Email:             getString(mapClaims, "email"),
		PreferredUsername: getString(mapClaims, "preferred_username"),
	}

	// Extract realm roles
	if realmAccess, ok := mapClaims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if r, ok := role.(string); ok {
					userClaims.RealmRoles = append(userClaims.RealmRoles, r)
				}
			}
		}
	}

	return userClaims, nil
}

func (j *JWTAuth) getPublicKey(kid string) (interface{}, error) {
	j.jwksMutex.RLock()
	jwks := j.jwks
	lastFetch := j.lastFetch
	j.jwksMutex.RUnlock()

	// Refresh JWKS every 5 minutes or if not fetched yet
	if jwks == nil || time.Since(lastFetch) > 5*time.Minute {
		if err := j.fetchJWKS(); err != nil {
			return nil, err
		}
		j.jwksMutex.RLock()
		jwks = j.jwks
		j.jwksMutex.RUnlock()
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return j.parseRSAPublicKey(key)
		}
	}

	// Key not found, try refreshing JWKS
	if err := j.fetchJWKS(); err != nil {
		return nil, err
	}

	j.jwksMutex.RLock()
	defer j.jwksMutex.RUnlock()

	for _, key := range j.jwks.Keys {
		if key.Kid == kid {
			return j.parseRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

func (j *JWTAuth) fetchJWKS() error {
	j.jwksMutex.Lock()
	defer j.jwksMutex.Unlock()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(j.keycloakConfig.JWKSURL())
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	j.jwks = &jwks
	j.lastFetch = time.Now()
	j.logger.Debug("JWKS fetched successfully", zap.Int("keys", len(jwks.Keys)))

	return nil
}

func (j *JWTAuth) parseRSAPublicKey(jwk JWK) (interface{}, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

func (j *JWTAuth) unauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "unauthorized",
		"message": message,
	})
}

func getString(m jwt.MapClaims, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func GetUserFromContext(ctx context.Context) *UserClaims {
	if user, ok := ctx.Value(UserContextKey).(*UserClaims); ok {
		return user
	}
	return nil
}
