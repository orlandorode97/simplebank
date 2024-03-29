package grpc

import (
	"github.com/orlandorode97/simple-bank/config"
	simplebankpb "github.com/orlandorode97/simple-bank/generated/simplebank"
	"github.com/orlandorode97/simple-bank/pkg/token"
	"github.com/orlandorode97/simple-bank/store"
	"github.com/orlandorode97/simple-bank/workers"
	"go.uber.org/zap"
)

type GRPCServer struct {
	simplebankpb.UnimplementedSimplebankServiceServer
	store           store.Store
	config          config.Config
	tokenMaker      token.Maker
	logger          *zap.SugaredLogger
	taskDistributor workers.TaskDistributor
}

func NewServer(conf config.Config, store store.Store, logger *zap.SugaredLogger, taskDistributor workers.TaskDistributor) (*GRPCServer, error) {
	tokenMaker, err := token.NewPasetoMaker(conf.SymmetricKey)
	if err != nil {
		return nil, err
	}

	return &GRPCServer{
		store:           store,
		config:          conf,
		tokenMaker:      tokenMaker,
		logger:          logger,
		taskDistributor: taskDistributor,
	}, nil
}
