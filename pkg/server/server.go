package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	accesspb "github.com/100mslive/access/pkg/internal"
	"github.com/100mslive/auth"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var (
	logger  *log.Logger
	grpcLog grpclog.LoggerV2
	// ErrDatabaseNotReady ...
	ErrDatabaseNotReady = errors.New("database not ready")
)

func init() {
	grpcLog = grpclog.NewLoggerV2(os.Stdout, os.Stderr, os.Stderr)
	grpclog.SetLoggerV2(grpcLog)
	logger = log.New(os.Stdout, "Access Server>", log.LstdFlags|log.Llongfile|log.Lmsgprefix)
}

// Config ...
type Config struct {
	Port     int    `mapstructure:"port,omitempty"`
	Cert     string `mapstructure:"cert,omitempty"`
	Key      string `mapstructure:"key,omitempty"`
	Metrics  int    `mapstructure:"metrics,omitempty"`
	Logging  bool   `mapstructure:"logging,omitempty"`
	Recovery bool   `mapstructure:"recovery,omitempty"`
	Caching  bool   `mapstructure:"caching,omitempty"`
}

// Server ...
type Server struct {
	connected bool
	service   string
	config    *Config
	health    *health.Server
	shutdown  chan struct{}
	auth      auth.Client
	accesspb.UnimplementedAccessServer
}

// New ...
func New(config *Config, options ...Option) *Server {
	logger.Println("Create server :", config)

	opts := newOptions()
	for _, opt := range options {
		opt(opts)
	}

	return &Server{config: config,
		health:   health.NewServer(),
		shutdown: make(chan struct{}),
		service:  "access",
		auth:     opts.auth,
	}
}

// DefaultConfig default config
func DefaultConfig() *Config {
	return &Config{
		Port:     8003,
		Metrics:  5053,
		Logging:  true,
		Recovery: true,
	}
}

func (config Config) String() string {
	return fmt.Sprintf("Port: %d", config.Port)
}

// Health check
func (server *Server) Health(ctx context.Context) error {
	if !server.connected {
		logger.Println(ErrServerNotConnected.Error())
		return ErrServerNotConnected
	}
	return nil
}

func (server *Server) watch() {
	ticker := time.NewTicker(time.Second * 5)
	server.health.SetServingStatus(server.service, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	go func() {
		for {
			select {
			case <-ticker.C:
				status := grpc_health_v1.HealthCheckResponse_SERVING
				if err := server.Health(context.TODO()); err != nil {
					status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
				}
				server.health.SetServingStatus(server.service, status)
			case <-server.shutdown:
				return
			}
		}
	}()
}

// Start start server
func (server *Server) Start(ctx context.Context) error {
	var options []grpc.ServerOption

	if server.config.Cert != "" && server.config.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(server.config.Cert, server.config.Key)
		if err != nil {
			logger.Println("Error!!!Failed to load cert:", err)
			return err
		}
		options = append(options, grpc.Creds(creds))

	}

	if err := server.auth.Connect(context.TODO()); err != nil {
		logger.Println("Error: ", err)
		return err
	}
	if err := server.auth.Ping(context.TODO()); err != nil {
		logger.Println("Error: ", err)
		return err
	}

	var streamInterceptor []grpc.StreamServerInterceptor
	var unaryInterceptor []grpc.UnaryServerInterceptor
	if server.config.Metrics > 0 {
		streamInterceptor = append(streamInterceptor, grpc_prometheus.StreamServerInterceptor)
		unaryInterceptor = append(unaryInterceptor, grpc_prometheus.UnaryServerInterceptor)
	}
	if server.config.Logging {
		grpcLoggingOpts := []grpc_logrus.Option{
			grpc_logrus.WithDecider(func(methodFullName string, err error) bool {
				// will not log gRPC calls if it was a call to healthcheck and no error was raised
				if err == nil && methodFullName == "/grpc.health.v1.Health/Check" {
					return false
				}
				// by default you will log all calls
				return true
			}),
		}
		streamInterceptor = append(streamInterceptor, grpc_ctxtags.StreamServerInterceptor())
		streamInterceptor = append(streamInterceptor, grpc_logrus.StreamServerInterceptor(logrus.NewEntry(logrus.New()), grpcLoggingOpts...))
		unaryInterceptor = append(unaryInterceptor, grpc_ctxtags.UnaryServerInterceptor())
		unaryInterceptor = append(unaryInterceptor, grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New()), grpcLoggingOpts...))
	}
	if server.config.Recovery {
		// Shared options for the logger, with a custom gRPC code to log level function.
		grpcRecoveryOpts := []grpc_recovery.Option{
			grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				return status.Errorf(codes.Unknown, "panic triggered: %v", p)
			}),
		}
		streamInterceptor = append(streamInterceptor, grpc_recovery.StreamServerInterceptor(grpcRecoveryOpts...))
		unaryInterceptor = append(unaryInterceptor, grpc_recovery.UnaryServerInterceptor(grpcRecoveryOpts...))
	}
	if len(streamInterceptor) > 0 {
		options = append(options, grpc_middleware.WithStreamServerChain(streamInterceptor...))
	}
	if len(unaryInterceptor) > 0 {
		options = append(options, grpc_middleware.WithUnaryServerChain(unaryInterceptor...))
	}

	grpcServer := grpc.NewServer(options...)

	server.watch()
	logger.Println("Start server on port", server.config.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", server.config.Port))
	if err != nil {
		logger.Println("Error!!!Failed to listen on port: ", err)
		return err
	}
	accesspb.RegisterAccessServer(grpcServer, server)

	grpc_health_v1.RegisterHealthServer(grpcServer, server.health)
	if server.config.Metrics > 0 {
		go func() {
			grpc_prometheus.Register(grpcServer)
			grpc_prometheus.EnableHandlingTimeHistogram()
			http.Handle("/metrics", promhttp.Handler())
			logger.Println("Starting metrics server : ", server.config.Metrics)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", server.config.Metrics), nil); err != nil {
				logger.Fatalf("Error in metrics server > %v", err)
			}
		}()
	}
	server.connected = true
	if err := grpcServer.Serve(lis); err != nil {
		logger.Println("Error!!!Failed to start server", err)
		return err
	}

	return nil
}
