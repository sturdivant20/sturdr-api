package api

import (
	"context"
	"database/sql"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sturdivant20/sturdr-api/include/navigation"
	"github.com/sturdivant20/sturdr-api/include/satellite"
	"github.com/sturdivant20/sturdr-api/include/telemetry"
)

type Application struct {
	cfg Config
	db  *sql.DB
}

// Init
func (app *Application) Init(cfg_fname string) error {
	var err error

	// 1. parse toml config file
	app.cfg, err = parseSettings(cfg_fname)
	if err != nil {
		return err
	}

	// 2. initialize database file
	app.db, err = initDatabase(app.cfg.Database.DbFile, app.cfg.Database.Clear)
	if err != nil {
		return err
	}

	return nil
}

// Mount
func (app *Application) Mount() http.Handler {
	// 1. create database tables/services
	sql := &app.cfg.Sql
	s_navigation := navigation.NewNavigationService(app.db, sql.NavigationCmds)
	h_navigation := navigation.NewHttpHandler(s_navigation)
	s_satellite := satellite.NewSatelliteService(app.db, sql.SatelliteCmds)
	h_satellite := satellite.NewHttpHandler(s_satellite)
	s_telemetry := telemetry.NewTelemetryService(app.db, sql.TelemetryCmds)
	h_telemetry := telemetry.NewHttpHandler(s_telemetry)

	// 2. create http endpoints
	ep := &app.cfg.Endpoints
	router := http.NewServeMux()
	router.HandleFunc(ep.Navigation+ep.Create, h_navigation.Create)
	router.HandleFunc(ep.Navigation+ep.Read, h_navigation.Read)
	router.HandleFunc(ep.Navigation+ep.Update, h_navigation.Update)
	router.HandleFunc(ep.Navigation+ep.Delete, h_navigation.Delete)

	router.HandleFunc(ep.Satellite+ep.Create, h_satellite.Create)
	router.HandleFunc(ep.Satellite+ep.Read, h_satellite.Read)
	router.HandleFunc(ep.Satellite+ep.Update, h_satellite.Update)
	router.HandleFunc(ep.Satellite+ep.Delete, h_satellite.Delete)

	router.HandleFunc(ep.Telemetry+ep.Create, h_telemetry.Create)
	router.HandleFunc(ep.Telemetry+ep.Read, h_telemetry.Read)
	router.HandleFunc(ep.Telemetry+ep.Update, h_telemetry.Update)
	router.HandleFunc(ep.Telemetry+ep.Delete, h_telemetry.Delete)

	// gui
	fs := http.FileServer(http.Dir("./gui"))
	router.Handle("/static/", http.StripPrefix("/static/", fs))
	router.HandleFunc(ep.Gui, func(w http.ResponseWriter, r *http.Request) {
		if ep.Gui == "/" && r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./gui/nav-view.html")
	})
	router.HandleFunc("/satellite-view", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./gui/sat-view.html")
	})
	router.HandleFunc("/guilog", remoteLogHandler)

	return router
}

// Run
func (app *Application) Run(ctx context.Context, router http.Handler) error {
	addr := app.cfg.Server.Host + ":" + strconv.Itoa(app.cfg.Server.Port)
	svr := &http.Server{
		Addr:         addr,
		Handler:      router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// run server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server has started at address '%s' ...", addr)
		serverErrors <- svr.ListenAndServe()
	}()

	// wait for termination signal or error
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		log.Println("Shutdown signal received ...")

		// 1. Set a deadline for the HTTP server to finish active requests
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 2. Stop the HTTP server first
		log.Println("Closing http server ...")
		if err := svr.Shutdown(shutdownCtx); err != nil {
			// If shutdown fails (timeout), force close the server
			svr.Close()
			log.Printf("HTTP shutdown error: %s", err.Error())
		}

		// 3. Now that no more requests are being processed, close the DB
		if app.db != nil {
			log.Println("Closing database connection ...")
			if err := app.db.Close(); err != nil {
				log.Printf("Database close error: %s", err.Error())
			}
		}

		log.Println("Graceful shutdown complete ...")
		return nil
	}
}

// // Sanitize trailing slashes in url
// func sanitizeSlashes(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// If the path is more than just "/" and ends in "/", trim it
// 		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
// 			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func remoteLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	// Read the log message from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading remote log: %v", err)
		return
	}
	defer r.Body.Close()

	// Print the message to your Go terminal
	log.Printf("[GUI] %s", string(body))

	// Respond with 200 OK
	w.WriteHeader(http.StatusOK)
}
