package router

import (
	"errors"

	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/config"
	internal_database "github.com/CloudDetail/apo/backend/internal/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	pkg_database "github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type resource struct {
	mux         *core.Mux
	logger      *zap.Logger
	ch          clickhouse.Repo
	prom        prometheus.Repo
	pol         polarisanalyzer.Repo
	internal_db internal_database.Repo
	pkg_db      pkg_database.Repo

	k8sApi kubernetes.Repo
}

type Server struct {
	Mux *core.Mux
}

func NewHTTPServer(logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	mux, err := core.New(logger)
	if err != nil {
		panic(err)
	}

	r := new(resource)
	r.logger = logger
	r.mux = mux

	// 初始化 Database
	dbRepo, err := internal_database.New(logger)
	if err != nil {
		logger.Fatal("new database err", zap.Error(err))
	}
	r.internal_db = dbRepo

	//初始化 sqlite
	pkgRepo, err := pkg_database.New(logger)
	if err != nil {
		logger.Fatal("new database err", zap.Error(err))
	}
	r.pkg_db = pkgRepo

	// 初始化 ClickHouse
	cfg := config.Get().ClickHouse
	chRepo, err := clickhouse.New(logger, []string{cfg.Address}, cfg.Database, cfg.Username, cfg.Password)
	if err != nil {
		logger.Fatal("new clickhouse err", zap.Error(err))
	}
	r.ch = chRepo

	// 初始化 Promethues
	promCfg := config.Get().Promethues
	promRepo, err := prometheus.New(logger, promCfg.Address, promCfg.Storage)
	if err != nil {
		logger.Fatal("new promethues err", zap.Error(err))
	}
	r.prom = promRepo

	// 初始化 PolarisAnalyzer
	polRepo, err := polarisanalyzer.New()
	if err != nil {
		logger.Fatal("new polarisanalyzer err", zap.Error(err))
	}
	r.pol = polRepo

	k8sCfg := config.Get().Kubernetes
	k8sApi, err := kubernetes.New(logger, k8sCfg.AuthType, k8sCfg.AuthFilePath)
	if err != nil {
		logger.Fatal("new kubernetes api err", zap.Error(err))
	}
	r.k8sApi = k8sApi

	// 设置 API 路由
	setApiRouter(r)

	s := new(Server)
	s.Mux = mux
	return s, nil
}
