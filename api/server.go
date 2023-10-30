package api

import (
	"github.com/gin-gonic/gin"
	"github.com/vpaklatzis/conduit-go/config"
	"github.com/vpaklatzis/conduit-go/logger"
)

type Server struct {
	conf   config.Config
	router *gin.Engine
	log    logger.Logger
}

func NewServer(conf config.Config, log logger.Logger) *Server {
	var engine *gin.Engine
	if conf.Environment == "test" {
		gin.SetMode(gin.ReleaseMode)
		log.Info("Test environment detected")
		engine = gin.New()
	} else {
		engine = gin.Default()
	}
	server := &Server{
		conf:   conf,
		router: engine,
		log:    log,
	}
	return server
}

func (s *Server) MountHandlers() {
	api := s.router.Group("/api")
	api.POST("/cves/search/keyword", s.KeywordHandler)
	api.POST("/scan", s.ScanHandler)
	api.POST("/rat", s.RatHandler)
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) Router() *gin.Engine {
	return s.router
}
