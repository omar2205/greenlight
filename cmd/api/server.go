package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log"
)

func (app *application) serve() error {
	srv := &http.Server {
		Addr: fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog: log.New(app.logger, "", 0),
	}

	// shutdown channel
	// receives any errors returned by the graceful Shutdown()
	shutdownError := make(chan error)

	// signal catcher
	go func() {
		// quit channel carries os.Signal values
		quit := make(chan os.Signal, 1)

		// notify listen to incomgin SIGINT and SIGTERM
		// relay them to quit channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel
		// * this code will block until a signal is received
		s := <-quit

		app.logger.PrintInfo("shutting down server", map[string]string {
			"signal": s.String(),
		})
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.PrintInfo("completing background tasks", nil)
		app.wg.Wait()

		shutdownError <- nil
	}()

	app.logger.PrintInfo("starting server", map[string]string {
		"addr": srv.Addr,
		"env": app.config.env,
	})

	// calling Shutdown() causes listen&serve to return ErrServerClosed
	// meaning graceful shutdown has started
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// otherwise, we wait to receive the return value from Shutdown()
	// if return value is an error, we know there was a problem with
	// the graceful shutdown
	err = <- shutdownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string {
		"addr": srv.Addr,
	})

	return nil
}