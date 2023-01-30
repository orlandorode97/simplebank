package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type serverHealthzResponse struct {
	Status  int64  `json:"status"`
	Message string `json:"message,omitempty"`
	Error   error  `json:"error,omitempty"`
}

func (s *Server) serverHealthz(ctx *gin.Context) {
	err := s.store.Ping()
	if err != nil {
		ctx.JSON(http.StatusBadGateway, serverHealthzResponse{
			Status: http.StatusBadGateway,
			Error:  err,
		})
		return
	}

	ctx.JSON(http.StatusOK, serverHealthzResponse{
		Status:  http.StatusOK,
		Message: "service is serving",
	})
}
