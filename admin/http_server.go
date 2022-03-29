package admin

import (
	"context"
	"github.com/0990/avatar-fight-server/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

type httpServer struct {
	listenAddr string
	router     *http.ServeMux
	server     *http.Server
}

func newHTTPServer(listenAddr string) *httpServer {
	s := &httpServer{}
	s.router = http.NewServeMux()
	s.listenAddr = listenAddr
	s.routes()
	return s
}

func (s *httpServer) Run() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	s.server = &http.Server{Handler: s.router}
	util.SafeGo(func() {
		err := s.server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			logrus.WithFields(logrus.Fields{
				"ListenAddr": s.listenAddr,
			}).WithError(err).Error("http.ListenAndServer error")
			panic(err)
		}
	})
	return nil
}

func (p *httpServer) routes() {
	p.router.Handle("/metrics", promhttp.Handler())
}

func (s *httpServer) GraceShutDown() error {
	//给服务端最多30秒的关闭时间，如果服务端还没关好，强制结束整个服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
