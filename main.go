package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robert-pkg/gateway/conf"
	"github.com/robert-pkg/gateway/server"
	"github.com/robert-pkg/micro-go/log"
	zap_log "github.com/robert-pkg/micro-go/log/zap-log"
	consul_registry "github.com/robert-pkg/micro-go/registry/consul"
	jaeger_trace "github.com/robert-pkg/micro-go/trace/jaeger-trace"
)

func main() {
	flag.Parse()

	// 初始化配置信息
	if err := conf.Init(); err != nil {
		panic(err)
	}

	if err := zap_log.InitByConfig(&conf.Conf.Log); err != nil {
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			log.Error("crash", "err", err)
		}

		log.Close()
	}()

	log.Info("start")

	_, tracerCloser, err := jaeger_trace.NewTracer("gateway", &conf.Conf.TraceConfig)
	if err != nil {
		panic(err)
	}
	defer tracerCloser.Close()

	// 根据系统配置，使用etcd， consul
	consul_registry.InitRegistryAsDefault(nil)

	svr := server.NewServer(conf.Conf)
	svr.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for {
		s := <-c
		//log.Infof("catch signal:%d", s)

		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			svr.Shutdown()
			time.Sleep(time.Second * 2)
			log.Info("exit.")
			return
		case syscall.SIGHUP:
			// TODO reload
		}
	}
}
