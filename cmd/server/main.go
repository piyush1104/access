package main

import (
	"context"

	accessServer "github.com/100mslive/access/pkg/server"
	auth "github.com/100mslive/auth/client"
	"github.com/100mslive/packages/conf"
	"github.com/100mslive/packages/log"
	"github.com/100mslive/packages/version"
)

var (
	// AppName ...
	AppName = "access"

	// AppDescription ...
	AppDescription = "100ms access Service"
)

func init() {
	// Init logging
	version.SetAppInfo(AppName, AppDescription)
}

func main() {
	// Print version info
	version.Print()
	// Load config
	config := conf.New(
		conf.WithConfigFileFlagName("config"),
		conf.WithConfigTypeFlagName("type"),
	)

	// auth
	authConfig := auth.DefaultConfig()
	config.Register("auth", authConfig)

	// server
	serverConfig := accessServer.DefaultConfig()
	config.Register("server", serverConfig)

	// log
	logConfig := log.DefaultConfig()
	config.Register("log", logConfig)

	if err := config.LoadWithFlag(); err != nil {
		log.Panicf("Failed to load config file %v", err)
	}

	// Init Log
	log.Init(logConfig)
	ctx := context.Background()

	server := accessServer.New(serverConfig, accessServer.WithAuth(auth.New(authConfig,
		auth.WithMetrics(true),
		auth.WithRetry(true),
		auth.WithTracing(true))))

	if err := server.Start(ctx); err != nil {
		log.Panicf("Failed to start server %v", err)
	}
}
