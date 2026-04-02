package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/flock"
	"github.com/f6o/qai/internal/gen/qaipb"
	"github.com/f6o/qai/internal/server"
	"github.com/f6o/qai/internal/service"
	"github.com/f6o/qai/internal/storage"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the QAI gRPC server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.EnsureDirectories(); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}

		socketPath := cfg.Server.SocketPath

		// Acquire lock to prevent multiple server instances
		lockPath := socketPath + ".lock"
		fl := flock.New(lockPath)
		locked, err := fl.TryLock()
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
		if !locked {
			return fmt.Errorf("another qai server is already running")
		}
		defer fl.Unlock()

		// Remove stale socket file
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove stale socket: %w", err)
		}

		// Ensure socket directory exists
		if err := os.MkdirAll(filepath.Dir(socketPath), 0755); err != nil {
			return fmt.Errorf("failed to create socket directory: %w", err)
		}

		lis, err := net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", socketPath, err)
		}
		defer lis.Close()

		// Set socket permissions to owner-only
		os.Chmod(socketPath, 0600)

		ts := storage.NewTaskStorage(cfg.Data.Todofile)
		ls := storage.NewLogStorage(cfg.Data.Logfile)
		taskSvc := service.NewLocalTaskService(ts)
		logSvc := service.NewLocalLogService(ls)

		grpcServer := grpc.NewServer()
		qaipb.RegisterQaiServiceServer(grpcServer, server.NewServer(taskSvc, logSvc))

		// Graceful shutdown on SIGINT/SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cmd.Println("\nShutting down server...")
			grpcServer.GracefulStop()
		}()

		cmd.Printf("QAI server listening on %s\n", socketPath)
		return grpcServer.Serve(lis)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
