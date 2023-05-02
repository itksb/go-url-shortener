package main

import (
	"context"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/app"
	"github.com/itksb/go-url-shortener/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// go run -ldflags "-X main.Version=v1.1.1 \
// -X 'main.buildVersion=1.1.1'" ./cmd/shortener/main.go
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {

	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion,
		buildDate,
		buildCommit,
	)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg.UseOsEnv()
	cfg.UseFlags()
	cfg.UseConfigFile()

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}

	////

	doneCh := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
		<-sigint
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
		defer cancelFunc()
		if err2 := application.HTTPServer.Shutdown(ctx); err2 != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err2)
		}
		close(doneCh)
	}()

	log.Println(application.Run())
	<-doneCh
	err = application.Close()
	if err != nil {
		log.Printf("error while closing application %s", err)
	}

	log.Println("bye bye")
}
