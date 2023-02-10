package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orlandorode97/simple-bank/pkg/token"
)

type AuthKey string

const (
	authorizationHeaderKey  = "authorization"
	authorizationHeaderType = "bearer"
)

var authorizationKey AuthKey = "auth_payload"

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey) // Get headers
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not presented in the request")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader) //Convert headers into []string
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authType := strings.ToLower(fields[0]) // Get the Bearer type
		if authType != authorizationHeaderType {
			err := errors.New("unsupported authorization header type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1] // Get access token

		payload, err := tokenMaker.VerfifyToken(accessToken) // Verify token
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(string(authorizationKey), payload) // Set the token payload in the context of the request.
		ctx.Next()                                 // Continue to the next handler
	}
}
