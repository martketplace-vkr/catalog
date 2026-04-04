package app

import (
	"context"

	trmsqlx "github.com/avito-tech/go-transaction-manager/sqlx"
	txmanager "github.com/avito-tech/go-transaction-manager/trm/manager"

	"github.com/martketplace-vkr/catalog/config"
	"github.com/martketplace-vkr/catalog/internal/app/cmp/server"
	adminRepository "github.com/martketplace-vkr/catalog/internal/repository/pg/admin"
	clientRepository "github.com/martketplace-vkr/catalog/internal/repository/pg/client"
	adminService "github.com/martketplace-vkr/catalog/internal/service/admin"
	clientService "github.com/martketplace-vkr/catalog/internal/service/client"
	adminTransport "github.com/martketplace-vkr/catalog/internal/transport/grpc/v1/admin"
	clientTransport "github.com/martketplace-vkr/catalog/internal/transport/grpc/v1/client"
	vendorTransport "github.com/martketplace-vkr/catalog/internal/transport/grpc/v1/vendor"

	"github.com/martketplace-vkr/pkg/build"
	"github.com/martketplace-vkr/pkg/build/components/pgxsqlxcomponent"
)

func Run(ctx context.Context, cfg *config.Config) error {
	pg := pgxsqlxcomponent.New(cfg.Postgres)

	txManager, err := txmanager.New(trmsqlx.NewDefaultFactory(pg.DB))
	if err != nil {
		return err
	}

	clientRepo := clientRepository.New(pg.DB, trmsqlx.DefaultCtxGetter)
	adminRepo := adminRepository.New(pg.DB, trmsqlx.DefaultCtxGetter)
	clientServ := clientService.New(txManager, clientRepo)
	adminServ := adminService.New(txManager, adminRepo)
	clientHandler := clientTransport.New(clientServ)
	vendorHandler := vendorTransport.New(clientServ)
	adminHandler := adminTransport.New(adminServ)

	grpcServer := server.New(
		cfg.Grpc,
		clientHandler,
		vendorHandler,
		adminHandler,
	)

	cmps := build.Components{
		pg,
		grpcServer,
	}

	app, err := build.NewApp(cmps)
	if err != nil {
		return err
	}

	return build.Run(ctx, app)
}
