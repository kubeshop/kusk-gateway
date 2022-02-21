package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kubeshop/kusk-gateway/internal/helper/httpserver"
	"github.com/kubeshop/kusk-gateway/internal/helper/management"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func main() {

	var (
		helperConfigurationManagerServiceAddress string
		log                                      *zap.SugaredLogger
		fleetID                                  string
	)
	flag.StringVar(&helperConfigurationManagerServiceAddress, "helper-config-manager-service-address", "", "The address (hostname:port) of Kusk Gateway Helper Configuration Manager Service")
	flag.StringVar(&fleetID, "fleetID", "", "The Envoy Fleet ID this Helper server is deployed for.")
	flag.Parse()
	log = initLogger().Sugar()
	defer log.Sync()

	nodeName, err := os.Hostname()
	if err != nil {
		log.Fatal("Cannot find out the local hostname")
	}
	log.Infof("Local node name: %s", nodeName)

	mainHandler := httpserver.NewMainHandler()
	httpServer := httpserver.NewHTTPServer(log.Desugar(), mainHandler)
	log.Infof("Starting the HTTP server on %s", httpServer.Addr)
	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{Time: 10 * time.Second, Timeout: 20 * time.Second, PermitWithoutStream: true},
		),
	}
	// Creates the connection, wait for the commands and respond
	// Using closure since this respawns if failed and we defer a lot of closing operations.
	connection := func() {
		log.Info("Dialing to the management service at: ", helperConfigurationManagerServiceAddress)
		conn, err := grpc.Dial(helperConfigurationManagerServiceAddress, grpcOpts...)
		defer conn.Close()
		if err != nil {
			log.Errorf("failed to dial: %v", err)
			return
		}
		client := management.NewConfigManagerClient(conn)
		// After 30 minutes the connection will restart
		// TODO: need to hide Error in the logs
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		stream, err := client.GetSnapshot(ctx, &management.ClientParams{NodeName: nodeName, FleetID: fleetID})
		if err != nil {
			log.Errorf("Got error during the connection to the management service: %s", err)
			return
		}
		// Receive snapshots and use them to update the configuration
		for {
			snapshot, err := stream.Recv()
			if err == io.EOF {
				log.Error("Got EOF during the receiving from the stream")
				break
			} else if err != nil {
				log.Errorf("Got error when receiving from the stream: %s", err)
				return
			}
			log.Info("Retrieved the configuration snapshot: ", snapshot)
			// Update HTTP Handler with the new mock config
			mockConfig := management.ProtoMockConfigToMockConfig(snapshot.GetMockConfig())
			mainHandler.SetMockConfig(mockConfig)
		}
	}
	// Endless loop while waiting for commands from the management server
	// Retry connection if broken in 1s.
	for {
		connection()
		// Retry the logic
		time.Sleep(1 * time.Second)
	}
}

func initLogger() *zap.Logger {
	// encoding: console, json
	zapCfg := zap.Config{Encoding: "console",
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: false,
	}
	zapCfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapCfg.EncoderConfig = zapcore.EncoderConfig{}
	zapCfg.OutputPaths = []string{"stdout"}
	zapCfg.ErrorOutputPaths = []string{"stderr"}
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapCfg.EncoderConfig.TimeKey = "time"
	zapCfg.EncoderConfig.MessageKey = "message"
	zapCfg.EncoderConfig.LevelKey = "severity"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	logger, err := zapCfg.Build()
	if err != nil {
		fmt.Println("Failure initialising logger:", err)
		os.Exit(1)
	}
	return logger
}
