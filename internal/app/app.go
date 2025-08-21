package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

	"github.com/Everest13/fin-aggregator-service/internal/config"
	"github.com/Everest13/fin-aggregator-service/internal/server"
	"github.com/Everest13/fin-aggregator-service/internal/server/handler"
	"github.com/Everest13/fin-aggregator-service/internal/service/bank"
	"github.com/Everest13/fin-aggregator-service/internal/service/category"
	"github.com/Everest13/fin-aggregator-service/internal/service/monzo"
	"github.com/Everest13/fin-aggregator-service/internal/service/transaction"
	"github.com/Everest13/fin-aggregator-service/internal/service/uploader"
	"github.com/Everest13/fin-aggregator-service/internal/service/user"
	"github.com/Everest13/fin-aggregator-service/internal/utils/logger"
)

type App struct {
	cfg                 *config.Config
	dBPool              *pgxpool.Pool
	server              *server.Server
	finAggregatorServer *handler.FinAggregatorServer
	monzoService        *monzo.Service
	transactionService  *transaction.Service
	bankService         *bank.Service
	userService         *user.Service
	categoryService     *category.Service
	uploaderService     *uploader.Service
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run() error {
	return a.server.Run()
}

func (a *App) initDeps(ctx context.Context) error {
	err := a.initConfig()
	if err != nil {
		return err
	}

	err = a.initDb(ctx)
	if err != nil {
		return err
	}

	err = a.initService(ctx)
	if err != nil {
		return err
	}

	a.initImpl()
	a.initServer()

	return nil
}

func (a *App) initConfig() error {
	err := config.LoadValues()
	if err != nil {
		logger.Error("failed to load config", err)
		return err
	}

	httpGracefulTimeout, err := time.ParseDuration(viper.GetString(config.HTTPGracefulTimeout))
	if err != nil {
		logger.Error("failed to parse HTTP graceful timeout", err)
		return fmt.Errorf("failed to parse http gracefull timeout duration config: %w", err)
	}

	httpClientTimeout, err := time.ParseDuration(viper.GetString(config.HTTPClientTimeout))
	if err != nil {
		logger.Error("failed to parse HTTP client timeout", err)
		return fmt.Errorf("failed to parse dBPool max conns lifetime duration config: %w", err)
	}

	dBMaxConLifetime, err := time.ParseDuration(viper.GetString(config.DBMaxConLifetime))
	if err != nil {
		logger.Error("failed to parse DB max conn lifetime", err)
		return fmt.Errorf("failed to parse dBPool max conns lifetime duration config: %w", err)
	}

	a.cfg = &config.Config{
		GRPC: config.GRPCConfig{
			Port:    viper.GetString(config.GRPCPort),
			Network: viper.GetString(config.GRPCNetwork),
		},
		HTTP: config.HTTPConfig{
			Port:            viper.GetString(config.HTTPPort),
			Host:            viper.GetString(config.HTTPHost),
			GracefulTimeout: httpGracefulTimeout,
			ClientTimeout:   httpClientTimeout,
		},
		Monzo: config.Monzo{
			ClientID:     viper.GetString(config.MonzoClientID),
			ClientSecret: viper.GetString(config.MonzoClientSecret),
			RedirectURI:  viper.GetString(config.MonzoRedirectURI),
		},
		DB: config.DBConfig{
			Name:           viper.GetString(config.DBName),
			User:           viper.GetString(config.DBUser),
			Password:       viper.GetString(config.DBPassword),
			Host:           viper.GetString(config.DBHost),
			Port:           viper.GetUint16(config.DBPort),
			SSLMode:        viper.GetString(config.SSLMode),
			MaxCons:        viper.GetInt32(config.DBMaxCons),
			MinCons:        viper.GetInt32(config.DBMinCons),
			MaxConLifetime: dBMaxConLifetime,
		},
	}

	return nil
}

func (a *App) initImpl() {
	a.finAggregatorServer = handler.NewFinAggregatorServer(
		a.transactionService,
		a.bankService,
		a.categoryService,
		a.userService,
		a.uploaderService,
		a.monzoService,
	)
}

func (a *App) initServer() {
	opts := &server.Opts{
		GrpcPort:        a.cfg.GRPC.Port,
		GrpcNetwork:     a.cfg.GRPC.Network,
		HttpPort:        a.cfg.HTTP.Port,
		HttpHost:        a.cfg.HTTP.Host,
		GracefulTimeout: a.cfg.HTTP.GracefulTimeout,
	}
	a.server = server.NewServer(a.finAggregatorServer, opts)
}

func (a *App) initDb(ctx context.Context) error {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		a.cfg.DB.User,
		a.cfg.DB.Password,
		a.cfg.DB.Host,
		a.cfg.DB.Port,
		a.cfg.DB.Name,
		a.cfg.DB.SSLMode,
	)
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logger.Error("failed to parse db config", err)
		return fmt.Errorf("failed to parse db config: %w", err)
	}

	cfg.MaxConns = a.cfg.DB.MaxCons
	cfg.MinConns = a.cfg.DB.MinCons
	cfg.MaxConnLifetime = a.cfg.DB.MaxConLifetime

	dbPool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		logger.Error("unable to create db pool", err)
		return fmt.Errorf("unable to create pool: %w", err)
	}

	if err = dbPool.Ping(ctx); err != nil {
		logger.Error("failed to ping db", err)
		return fmt.Errorf("failed to ping db: %w", err)
	}

	a.dBPool = dbPool

	logger.Info(
		"Database connection established",
		a.cfg.DB.Host,
		strconv.Itoa(int(a.cfg.DB.Port)))

	return nil
}

func (a *App) initService(ctx context.Context) error {
	a.bankService = bank.NewService(a.dBPool)
	a.userService = user.NewService(a.dBPool)

	a.categoryService = category.NewService(a.dBPool)
	err := a.categoryService.Initialize(ctx)
	if err != nil {
		logger.Error("failed to initialize category service store", err)
		return err
	}

	a.transactionService = transaction.NewService(a.dBPool, a.categoryService)
	err = a.transactionService.Initialize(ctx)
	if err != nil {
		logger.Error("failed to initialize transaction service stores", err)
		return err
	}

	a.uploaderService = uploader.NewService(a.dBPool, a.bankService, a.transactionService, a.categoryService)
	err = a.uploaderService.Initialize(ctx)
	if err != nil {
		logger.Error("failed to initialize csv process service stores", err)
		return err
	}

	a.monzoService = monzo.NewService(&monzo.MonzoCfg{
		ClientID:     a.cfg.Monzo.ClientID,
		ClientSecret: a.cfg.Monzo.ClientSecret,
		RedirectURI:  a.cfg.Monzo.RedirectURI,
	},
		a.cfg.HTTP.ClientTimeout,
		a.transactionService,
		a.categoryService,
	)

	return nil
}
