package server

import (
	"fmt"
	"net"
	"net/http"
	"open-indexer/config"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

type ServerType int32

var SubHandler *Subscriber

const (
	HTTP ServerType = 0
	WS   ServerType = 1
)

type Server struct {
	srv    *rpc.Server
	stype  ServerType
	rpcAPI []rpc.API
	logger *logrus.Logger
}

func NewServer(rpcAPI []rpc.API, logger *logrus.Logger, stype ServerType) *Server {
	return &Server{
		srv:    rpc.NewServer(),
		stype:  stype,
		rpcAPI: rpcAPI,
		logger: logger,
	}
}

func (s *Server) Start() {

	for _, api := range s.rpcAPI {
		if err := s.srv.RegisterName(api.Namespace, api.Service); err != nil {
			s.logger.Fatalf("Could not register API: %w", err)
		}
	}

	var (
		handler http.Handler
		ser     string
		host    string
		port    int
	)

	s.logger.Info(config.Global.Netcfg.HttpHost, ":", config.Global.Netcfg.HTTPPort)
	s.logger.Info(config.Global.Netcfg.WSHost, ":", config.Global.Netcfg.WSPort)

	if s.stype == HTTP {
		// vhosts := splitAndTrim("http.vhost")
		// cors := splitAndTrim("http.corsdomain")
		// handler = node.NewHTTPHandlerStack(s.srv, cors, vhosts)
		handler = s.srv
		host = config.Global.Netcfg.HttpHost
		port = config.Global.Netcfg.HTTPPort
		ser = "http"

	}

	if s.stype == WS {
		SubHandler = NewSubscriber()
		wscors := splitAndTrim("*")
		handler = s.srv.WebsocketHandler(wscors)
		host = config.Global.Netcfg.WSHost
		port = config.Global.Netcfg.WSPort
		ser = "ws"
	}

	httpEndpoint := fmt.Sprintf("%s:%d", host, port)
	_, addr, err := s.startHTTPEndpoint(httpEndpoint, rpc.DefaultHTTPTimeouts, handler)
	if err != nil {
		s.logger.Fatalf("Could not start RPC api: %v", err)
	}

	extapiURL := fmt.Sprintf("%s://%v/", ser, addr)
	s.logger.Printf("%s endpoint opened url : %s", ser, extapiURL)
}

func (s *Server) Stop() {
	s.srv.Stop()
}

func (s *Server) startHTTPEndpoint(endpoint string, timeouts rpc.HTTPTimeouts, handler http.Handler) (*http.Server, net.Addr, error) {
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", endpoint); err != nil {
		return nil, nil, err
	}
	httpSrv := &http.Server{
		Handler:      handler,
		ReadTimeout:  timeouts.ReadTimeout,
		WriteTimeout: timeouts.WriteTimeout,
		IdleTimeout:  timeouts.IdleTimeout,
	}
	go httpSrv.Serve(listener)
	return httpSrv, listener.Addr(), err
}

func splitAndTrim(input string) (ret []string) {
	l := strings.Split(input, ",")
	for _, r := range l {
		if r = strings.TrimSpace(r); r != "" {
			ret = append(ret, r)
		}
	}
	return ret
}
