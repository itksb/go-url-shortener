package app

import (
	"errors"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/dbstorage"
	"github.com/itksb/go-url-shortener/internal/filestorage"
	"github.com/itksb/go-url-shortener/internal/grpcserver"
	"github.com/itksb/go-url-shortener/internal/grpcserver/interceptor"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/internal/router"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/internal/storage"
	"github.com/itksb/go-url-shortener/migrate"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/itksb/go-url-shortener/pkg/session"
	go_url_shortener "github.com/itksb/go-url-shortener/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

// App - application
type App struct {
	HTTPServer    *http.Server
	GRPCServer    *grpc.Server
	logger        logger.Interface
	urlshortener  *shortener.Service
	reposhortener shortener.ShortenerStorage
	enableHTTPS   bool

	grpcAddr string

	io.Closer
}

// NewApp - constructor of the App
func NewApp(cfg config.Config) (*App, error) {
	l, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	var repo shortener.ShortenerStorage
	var db *dbstorage.Storage

	if cfg.Dsn != "" { // use postgres database as the storage driver
		// run migrations
		err = migrate.Migrate(cfg.Dsn, migrate.Migrations)
		if err != nil {
			l.Error(fmt.Sprintf("migration error: %s", err.Error()))
			return nil, err
		}
		db, err = dbstorage.NewPostgres(cfg.Dsn, l, nil)
		if err != nil {
			l.Error(fmt.Sprintf("dbstorage.NewPostgres error: %s", err.Error()))
		}
		repo = db // pointer nothing criminal
	} else if cfg.FileStoragePath != "" {
		// file-based storage
		repo, err = filestorage.NewStorage(l, cfg.FileStoragePath)
		if err != nil {
			l.Error(fmt.Sprintf("File storage error: %s", err.Error()))
			return nil, err
		}
	} else {
		// inMemory storage
		repo = storage.NewStorage(l)
	}
	urlshortener := shortener.NewShortener(l, repo)

	h := handler.NewHandler(l, urlshortener, db, db, cfg)

	codec, err := session.NewSecureCookie([]byte(cfg.SessionConfig.HashKey), []byte(cfg.SessionConfig.BlockKey))
	if err != nil {
		l.Error(fmt.Sprintf("Codec for session creating error: %s", err.Error()))
		return nil, err
	}
	sessionStore := session.NewCookieStore(codec)

	routeHandler, err := router.NewRouter(
		h,
		sessionStore,
		l,
		cfg.Debug,
		cfg.TrustedSubnet,
	)
	if err != nil {
		l.Error(fmt.Sprintf("Router creating error: %s", err.Error()))
		return nil, err
	}

	srv := createHTTPServer(routeHandler, cfg)
	grpcSrv := createGRPCServer(
		cfg,
		l,
		db,
		urlshortener,
		codec,
	)

	l.Info("is debug environment? ", cfg.Debug)

	return &App{
		HTTPServer:    srv,
		GRPCServer:    grpcSrv,
		logger:        l,
		urlshortener:  urlshortener,
		reposhortener: repo,
		enableHTTPS:   cfg.EnableHTTPS,
		grpcAddr:      cfg.GRPCAddr,
		Closer:        nil,
	}, nil
}

// Run - run the application instance
func (app *App) Run() error {
	app.logger.Info("server starting", "addr", app.HTTPServer.Addr)
	g := errgroup.Group{}
	if app.enableHTTPS {
		g.Go(func() error {
			return app.HTTPServer.ListenAndServeTLS("", "")
		})
	} else {
		g.Go(app.HTTPServer.ListenAndServe)
	}

	g.Go(func() error {
		listen, err := net.Listen("tcp", app.grpcAddr)
		if err != nil {
			app.logger.Error("failed to listen grpsAddr", "error", err)
			return err
		}
		return app.GRPCServer.Serve(listen)
	})

	return g.Wait()
}

// Close -
func (app *App) Close() error {
	repoErr := app.reposhortener.Close()
	urlsErr := app.urlshortener.Close()

	msg := ""
	if repoErr != nil {
		msg = repoErr.Error()
	}
	if urlsErr != nil {
		msg = fmt.Sprintf("%s%s", msg, urlsErr.Error())
	}

	if len(msg) > 0 {
		return errors.New(msg)
	}
	return nil
}

func createHTTPServer(routeHandler http.Handler, cfg config.Config) *http.Server {
	var srv *http.Server

	if cfg.EnableHTTPS {
		// конструируем менеджер TLS-сертификатов
		manager := &autocert.Manager{
			// функция, принимающая Terms of Service издателя сертификатов
			Prompt: autocert.AcceptTOS,
		}
		// конструируем сервер с поддержкой TLS
		srv = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort),
			Handler: routeHandler,
			WriteTimeout: func() time.Duration {
				if cfg.Debug {
					return 0
				}
				return 15 * time.Second
			}(),
			// для TLS-конфигурации используем менеджер сертификатов
			TLSConfig: manager.TLSConfig(),
		}
	} else {
		srv = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort),
			Handler: routeHandler,
			WriteTimeout: func() time.Duration {
				if cfg.Debug {
					return 0
				}
				return 15 * time.Second
			}(),
		}
	}

	return srv
}

func createGRPCServer(
	cfg config.Config,
	l logger.Interface,
	dbping handler.IPingableDB,
	urlshortener *shortener.Service,
	codec session.Codec,
) *grpc.Server {
	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.CreateUnaryAuthInterceptor(codec),
		),
	)

	appServer := grpcserver.NewGRPCServer(
		dbping,
		l,
		urlshortener,
		cfg,
	)
	// регистрируем сервис
	go_url_shortener.RegisterShortenerServer(s, appServer)

	return s
}
