package client

import (
	"context"
	"log"
	"os"
	"sync/atomic"
	"time"

	accesspb "github.com/100mslive/access/pkg/internal"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

const (
	disconnected = 0
	connected    = 1
)

var (
	logger *log.Logger
	// DefaultDialTimeout  ...
	DefaultDialTimeout = 5000
)

func init() {
	logger = log.New(os.Stdout, "RPCPolicy>", log.LstdFlags|log.Llongfile|log.Lmsgprefix)
}

// Client ...
type Client struct {
	connection     int32
	metricsEnabled bool
	retryEnabled   bool
	config         *Config
	logger         *log.Logger
	conn           *grpc.ClientConn
	rpc            accesspb.AccessClient
	dialTimeout    time.Duration
}

// New ...
func New(config *Config, options ...Option) *Client {
	config.SetDefaults()
	if config.DialTimeout < 1000 {
		config.DialTimeout = DefaultDialTimeout
	}
	c := &Client{
		config:      config,
		logger:      logger,
		dialTimeout: time.Millisecond * time.Duration(config.DialTimeout),
	}

	for _, option := range options {
		option(c)
	}
	return c
}

// Connect ...
func (client *Client) Connect(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&client.connection, disconnected, connected) {
		return ErrClientNotConnected
	}
	// client.config.Addr = "localhost:9001"
	client.logger.Print("Connecting to => ", client.config.Addr)
	options := []grpc.DialOption{grpc.WithBlock()}
	if client.config.Cert == "" {
		options = append(options, grpc.WithInsecure())
	} else {
		creds, err := credentials.NewClientTLSFromFile(client.config.Cert, "")
		if err != nil {
			client.logger.Print("Error!!!", err)
			return err
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	}
	var streamInterceptor []grpc.StreamClientInterceptor
	var unaryInterceptor []grpc.UnaryClientInterceptor

	if client.metricsEnabled {
		options = append(options, grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor))
		options = append(options, grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	}
	if client.retryEnabled {
		opts := []grpc_retry.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
			grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.Unavailable),
			grpc_retry.WithMax(3),
		}
		streamInterceptor = append(streamInterceptor, grpc_retry.StreamClientInterceptor(opts...))
		unaryInterceptor = append(unaryInterceptor, grpc_retry.UnaryClientInterceptor(opts...))
	}
	if len(streamInterceptor) > 0 {
		options = append(options, grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptor...)))
	}
	if len(unaryInterceptor) > 0 {
		options = append(options, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptor...)))
	}
	ctx, cancel := context.WithTimeout(ctx, client.dialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, client.config.Addr, options...)
	if err != nil {
		client.logger.Println("Error!!!", err)
		return err
	}
	client.conn = conn
	client.rpc = accesspb.NewAccessClient(conn)
	client.logger.Print("Connection successfull")
	return nil
}

// Connected client connected status
func (client *Client) Connected() bool {
	return atomic.LoadInt32(&client.connection) == connected
}

// Close ...
func (client *Client) Close() error {
	if !atomic.CompareAndSwapInt32(&client.connection, connected, disconnected) {
		return ErrClientNotConnected
	}
	return nil
}

// Health ...
func (client *Client) Health(ctx context.Context) error {
	if !client.Connected() {
		return ErrClientNotConnected
	}
	return nil
}
