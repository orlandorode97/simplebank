package api

import (
	"github.com/gin-gonic/gin"
	"github.com/orlandorode97/simple-bank/config"
	"github.com/orlandorode97/simple-bank/pkg/token"
	"github.com/orlandorode97/simple-bank/store"
)

// Server serves http requests, routes, config, token generation, and gRPC calls.
type Server struct {
	store      store.Store
	handler    *gin.Engine
	config     config.Config
	tokenMaker token.Maker
}

func NewServer(conf config.Config, store store.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(conf.SymmetricKey)
	if err != nil {
		return nil, err
	}
	server := &Server{
		store:      store,
		config:     conf,
		tokenMaker: tokenMaker,
	}

	router := gin.New()

	v1 := router.Group("/api/v1")

	v1.POST("/login", server.login)
	v1.POST("/refresh_token", server.refreshAccessToken)
	v1.GET("/healthz", server.serverHealthz)
	server.addUserRoutes(v1)

	v1.Use(authMiddleware(tokenMaker))

	server.addAccountRoutes(v1)
	server.addTransferRoutes(v1)

	server.handler = router

	return server, err
}

func (s *Server) Listen(addr string) error {
	return s.handler.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
