package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/robert-pkg/gateway/client"
	"github.com/robert-pkg/gateway/conf"
	"github.com/robert-pkg/micro-go/log"
)

// Server .
type Server struct {
	cfg        *conf.Config
	dispatcher dispatcher
	httpSvr    *http.Server
}

// NewServer create server
func NewServer(cfg *conf.Config) *Server {

	return &Server{
		cfg: cfg,
	}
}

// Start 启动服务器
func (s *Server) Start() {

	if err := s.dispatcher.initAuth(s.cfg); err != nil {
		panic(err)
	}

	s.startHTTPServer()

}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {

	// 关闭 http 服务器(优雅退出)
	if s.httpSvr != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpSvr.Shutdown(ctx); err != nil {
			log.Error("Server shutdown error", "err", err)
		}
	}

	client.Stop()
}

func (s *Server) startHTTPServer() {

	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/api/:service/:method", s.httpHandler)

	//没有路由的页面
	//为没有配置处理函数的路由添加处理程序，默认情况下它返回404
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Not Found!")
	})

	s.httpSvr = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.ServerConfig.Port),
		Handler: r,
	}

	log.Info("start listen", "port", s.cfg.ServerConfig.Port)
	go func() {
		if err := s.httpSvr.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

}

func (s *Server) httpHandler(c *gin.Context) {

	servicename, methodname := c.Param("service"), c.Param("method")

	// read body
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Warn("warn", "warn", err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	//log.Info("new request ", "service_name", servicename, "methodname", methodname, "body", string(bodyBytes))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel() // terminal routine tree

	realServiceName := "community.interface." + servicename
	httpCode, rspStr := s.dispatcher.Call(ctx, c.Request.Header, realServiceName, methodname, bodyBytes)

	c.Header("Content-Type", "application/json")

	//log.Info("responce", "service_name", servicename, "methodname", methodname, "httpCode", httpCode, "rspStr", rspStr)

	c.String(httpCode, rspStr)
}
