package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/jwt"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/httputils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/jsonutil"
)

type userServiceForAuth interface {
	Register(ctx context.Context, login string, password string) (*domain.User, error)
	Auth(ctx context.Context, login string, password string) (*domain.User, error)
}

type AuthHandler struct {
	userService userServiceForAuth
}

func NewAuthHandler(userService userServiceForAuth) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

type registerBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (b *registerBody) Valid() bool {
	if len(b.Login) < 4 || len(b.Password) < 6 {
		return false
	}

	return true
}

type registerResponse struct {
	AccessToken string `json:"access_token"`
}

// Register godoc
// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param dto body registerBody true "Register new user"
// @Success 200 {object} registerResponse
// @Failure 400 {object} httputils.HTTPError
// @Failure 409 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody

	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	if !body.Valid() {
		httputils.SendJSONErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	user, err := h.userService.Register(r.Context(), body.Login, body.Password)

	if err != nil {
		if errors.Is(err, domain.ErrLoginAlreadyTaken) {
			httputils.SendJSONErrorResponse(w, http.StatusConflict, err.Error())
			return
		}

		log.Println(err)
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	accessToken, err := jwt.GenerateToken(user.ID)

	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not generate token")
		return
	}

	response := registerResponse{
		AccessToken: accessToken,
	}

	w.Header().Set("Authorization", "Bearer "+accessToken)

	httputils.SendJSONResponse(w, http.StatusOK, response)
}

type loginBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (b *loginBody) Valid() bool {
	if len(b.Login) == 0 || len(b.Password) == 0 {
		return false
	}

	return true
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

// Login godoc
// @Summary Login to account by credentials
// @Tags auth
// @Accept json
// @Produce json
// @Param dto body loginBody true "Login to account"
// @Success 200 {object} loginResponse
// @Failure 400 {object} httputils.HTTPError
// @Failure 401 {object} httputils.HTTPError
// @Failure 500 {object} httputils.HTTPError
// @Router /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if status, err := jsonutil.Unmarshal(w, r, &body); err != nil {
		httputils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	if !body.Valid() {
		httputils.SendJSONErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	user, err := h.userService.Auth(r.Context(), body.Login, body.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLoginOrPassword) {
			httputils.SendJSONErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, domain.ErrInternalServer.Error())
		return
	}

	accessToken, err := jwt.GenerateToken(user.ID)
	if err != nil {
		httputils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not generate token")
		return
	}

	response := loginResponse{
		AccessToken: accessToken,
	}

	w.Header().Set("Authorization", "Bearer "+accessToken)

	httputils.SendJSONResponse(w, http.StatusOK, response)
}
