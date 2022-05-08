package main

import (
	"context"
	"fmt"
	"github.com/djcass44/go-utils/logging"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	whhttp "github.com/slok/kubewebhook/v2/pkg/http"
	"github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
	"gitlab.com/autokubeops/kube-image-webhook/internal/logr"
	"gitlab.com/autokubeops/kube-image-webhook/internal/webhook"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type environment struct {
	Port int `envconfig:"PORT" default:"8443"`

	ConfigPath string `split_words:"true"`

	Log struct {
		Level int `split_words:"true"`
	}

	TLS struct {
		Cert string `split_words:"true" required:"true"`
		Key  string `split_words:"true" required:"true"`
	}
}

func main() {
	var e environment
	if err := envconfig.Process("webhook", &e); err != nil {
		stdlog.Fatal("failed to read environment")
		return
	}
	zc := zap.NewProductionConfig()
	zc.Level = zap.NewAtomicLevelAt(zapcore.Level(e.Log.Level * -1))
	log, ctx := logging.NewZap(context.TODO(), zc)

	whlogger := logr.NewLogger(log)

	// load config
	conf, err := config.Get(ctx, e.ConfigPath)
	if err != nil {
		log.Error(err, "failed to read config")
		os.Exit(1)
		return
	}

	// setup services
	svc := webhook.NewImageWebhook(log, conf)
	wh, err := mutating.NewWebhook(mutating.WebhookConfig{
		ID:      "kube-image-mutate",
		Mutator: svc,
		Obj:     &corev1.Pod{},
		Logger:  whlogger,
	})
	if err != nil {
		log.Error(err, "failed to setup webhook")
		os.Exit(1)
		return
	}
	// create a http handler
	handler, err := whhttp.HandlerFor(whhttp.HandlerConfig{
		Webhook: wh,
		Logger:  whlogger,
	})
	if err != nil {
		log.Error(err, "failed to setup webhook handler")
		os.Exit(1)
		return
	}

	// setup routing
	router := mux.NewRouter()
	router.Handle("/mutate", handler)
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	})

	// start the server
	go func() {
		addr := fmt.Sprintf(":%d", e.Port)
		log.Info("starting http server", "Addr", addr)
		if err := http.ListenAndServeTLS(addr, e.TLS.Cert, e.TLS.Key, router); err != nil {
			log.Error(err, "http server exited")
			os.Exit(1)
		}
	}()

	// wait for a signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigC
	log.Info("received SIGTERM/SIGINT (%s), shutting down...", "Signal", sig)
}
