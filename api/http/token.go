package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrBlockedSession = errors.New("blocked session")
	ErrInvalidSession = errors.New("invalid session user")
	ErrExpiredSession = errors.New("expired session")
)

type refreshAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) refreshAccessToken(c *gin.Context) {
	var req refreshAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	refreshPayload, err := s.tokenMaker.VerfifyToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := s.store.GetSession(c, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if session.IsBlocked {
		c.JSON(http.StatusUnauthorized, errorResponse(ErrBlockedSession))
		return
	}

	if session.Username != refreshPayload.Username {
		c.JSON(http.StatusUnauthorized, errorResponse(ErrInvalidSession))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, errorResponse(ErrExpiredSession))
		return
	}

	token, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Username, s.config.TokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := &refreshAccessTokenResponse{
		AccessToken:          token,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	c.JSON(http.StatusOK, resp)
}
