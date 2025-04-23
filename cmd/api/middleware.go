package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tenteedee/gopher-social/internal/store"
)

type userKey string

const userContextKey userKey = "user"

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("auth header is missing"))
				return
			}

			// parse the header -> base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid auth header"))
				return
			}

			// decode the header
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			password := app.config.auth.basic.password

			//check the credentials
			credentials := strings.SplitN(string(decoded), ":", 2)
			if len(credentials) != 2 || credentials[0] != username || credentials[1] != password {
				app.unauthorizedBasicErrorResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorized(w, r, fmt.Errorf("auth header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorized(w, r, fmt.Errorf("invalid auth header"))
			return
		}

		token := parts[1]
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorized(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorized(w, r, err)
			return
		}

		// if r.URL.Path == "/v1/users/me" {
		// 	r, err = app.attachUserToContext(r, userID)
		// 	if err != nil {
		// 		app.Unauthorized(w, r, err)
		// 		return
		// 	}

		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		r, err = app.attachUserToContext(r, userID)
		if err != nil {
			app.unauthorized(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) attachUserToContext(r *http.Request, userID int64) (*http.Request, error) {
	user, err := app.getUser(r.Context(), userID)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, userContextKey, user)

	return r.WithContext(ctx), nil
}

func (app *application) CheckPostOwnership(roles string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post, err := getPostFromContext(r)
		if err != nil {
			app.notFound(w, r, err)
			return
		}

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, roles)
		if err != nil {
			app.internalServerError(w, r, fmt.Errorf("invalid role: %w", err))
			return
		}

		if !allowed {
			app.forbidden(w, r, fmt.Errorf("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userId int64) (*store.User, error) {
	if !app.config.redisCfg.enabled {
		return app.store.User.GetById(ctx, userId)
	}

	app.logger.Infow(
		"cache hit",
		"key", "user",
		"id", userId,
	)

	user, err := app.cacheStorage.User.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		app.logger.Infow(
			"cache miss, fetching from db",
			"key", "user",
			"id", userId,
		)
		user, err = app.store.User.GetById(ctx, userId)
		if err != nil {
			return nil, err
		}

		if err := app.cacheStorage.User.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
