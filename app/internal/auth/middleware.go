package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Middleware struct {
	verifier    *oidc.IDTokenVerifier
	userIDClaim string
}

func NewMiddleware(
	ctx context.Context,
	issuerURL string,
	clientID string,
	userIDClaim string,
) (*Middleware, error) {
	issuerURL = strings.TrimSpace(issuerURL)
	if issuerURL == "" {
		return nil, errors.New("OIDC_ISSUER_URL is required")
	}

	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, errors.New("OIDC_CLIENT_ID is required")
	}

	userIDClaim = strings.TrimSpace(userIDClaim)
	if userIDClaim == "" {
		userIDClaim = "sub"
	}

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("create oidc provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return &Middleware{
		verifier:    verifier,
		userIDClaim: userIDClaim,
	}, nil
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		tokenString, err := bearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, err.Error())
			return
		}

		idToken, err := m.verifier.Verify(r.Context(), tokenString)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		userID, err := m.extractUserID(idToken)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, err.Error())
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithUserID(r.Context(), userID)))
	})
}

func (m *Middleware) extractUserID(idToken *oidc.IDToken) (string, error) {
	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("read token claims: %w", err)
	}

	value, ok := claims[m.userIDClaim]
	if !ok {
		return "", fmt.Errorf("token claim %s not found", m.userIDClaim)
	}

	userID, ok := value.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", fmt.Errorf("token claim %s must be a non-empty string", m.userIDClaim)
	}

	return userID, nil
}

func bearerToken(authorizationHeader string) (string, error) {
	authorizationHeader = strings.TrimSpace(authorizationHeader)
	if authorizationHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.SplitN(authorizationHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization header must use Bearer token")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("bearer token is required")
	}

	return token, nil
}

func writeAuthError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
