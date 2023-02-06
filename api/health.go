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

	/*
	   - Llenar telefono particular
	   - Fecha de ingreso mm-dd-yyyy
	   - Correo personal
	   - antiguedad 1 año y 2 meses que estuve en Wizeline
	   - Solicitar carta de antiguedad de mis seguros gastos mayores.
	   - Reglar para determinar la suma asegurada (no llenar)
	   - Salario bruto
	   - Guadalajara, Jal 30 enero 2023
	*/
}
