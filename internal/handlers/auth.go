package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/MowlCoder/accumulative-loyalty-system/internal/domain"
	"github.com/MowlCoder/accumulative-loyalty-system/internal/jwt"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/http_utils"
	"github.com/MowlCoder/accumulative-loyalty-system/pkg/json_util"
)

type userServiceForAuth interface {
	Register(ctx context.Context, login string, password string) (*domain.User, error)
	Auth(ctx context.Context, login string, password string) (*domain.User, error)
}

type AuthHandler struct {
	userService userServiceForAuth
}

type AuthHandlerOptions struct {
	UserService userServiceForAuth
}

func NewAuthHandler(options *AuthHandlerOptions) *AuthHandler {
	return &AuthHandler{
		userService: options.UserService,
	}
}

type registerBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type registerResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody

	if status, err := json_util.Unmarshal(w, r, &body); err != nil {
		http_utils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), body.Login, body.Password)

	if err != nil {
		if errors.Is(err, domain.ErrLoginAlreadyTaken) {
			http_utils.SendJSONErrorResponse(w, http.StatusConflict, err.Error())
			return
		}

		http_utils.SendJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, err := jwt.GenerateToken(user.ID)

	if err != nil {
		http_utils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not generate token")
		return
	}

	response := registerResponse{
		AccessToken: accessToken,
	}

	http_utils.SendJSONResponse(w, http.StatusOK, response)
}

type loginBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody

	if status, err := json_util.Unmarshal(w, r, &body); err != nil {
		http_utils.SendJSONErrorResponse(w, status, err.Error())
		return
	}

	user, err := h.userService.Auth(r.Context(), body.Login, body.Password)

	if err != nil {
		if errors.Is(err, domain.ErrInvalidLoginOrPassword) {
			http_utils.SendJSONErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		http_utils.SendJSONErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, err := jwt.GenerateToken(user.ID)

	if err != nil {
		http_utils.SendJSONErrorResponse(w, http.StatusInternalServerError, "can not generate token")
		return
	}

	response := loginResponse{
		AccessToken: accessToken,
	}

	http_utils.SendJSONResponse(w, http.StatusOK, response)
}
