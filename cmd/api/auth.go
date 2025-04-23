package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tenteedee/gopher-social/internal/mailer"
	"github.com/tenteedee/gopher-social/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// Register User godoc
//
//	@Summary		Register a user
//	@Description	Register a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/authentication/user [post]
func (app *application) registerUserhandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: store.Role{
			Name: "user",
		},
	}

	// hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashedToken := hex.EncodeToString(hash[:])

	// store the user
	err := app.store.User.CreateAndInvite(r.Context(), user, hashedToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrorDuplicateEmail:
			app.badRequest(w, r, err)
			return
		case store.ErrorDuplicateUsername:
			app.badRequest(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: fmt.Sprintf("%s/activate/%s", app.config.frontendURL, plainToken),
	}

	// send activation email to user
	statusCode, err := app.mailer.Send(
		mailer.UserWelcomeTemplate,
		user.Username,
		user.Email,
		vars,
		!isProdEnv,
	)
	if err != nil {
		app.logger.Errorw("failed to send activation email",
			"error", err,
		)
		//rollback if seding email fails
		if err := app.store.User.Delete(r.Context(), user.ID); err != nil {
			app.logger.Errorw("failed to delete user",
				"error", err,
			)
		}

		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "statusCode", statusCode)

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}

}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse payload credentials
	var payload CreateUserTokenPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequest(w, r, err)
		return
	}

	// check if user exists from the payload
	user, err := app.store.User.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.unauthorized(w, r, err) // not use NotFound due to Enumeration Attack
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	// check if password is correct
	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorized(w, r, fmt.Errorf("invalid password"))
		return
	}

	// generate token
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.issuer,
		"aud": app.config.auth.token.audience,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
