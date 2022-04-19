package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	whhttp "github.com/slok/kubewebhook/v2/pkg/http"
	"github.com/slok/kubewebhook/v2/pkg/log/logrus"
	"github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
	"gitlab.com/autokubeops/kube-image-webhook/internal/webhook"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type environment struct {
	Port int `envconfig:"PORT" default:"8080"`

	ConfigPath string `split_words:"true"`

	TLS struct {
		Cert string `split_words:"true" required:"true"`
		Key  string `split_words:"true" required:"true"`
	}
}

func main() {
	var e environment
	if err := envconfig.Process("webhook", &e); err != nil {
		log.WithError(err).Fatal("failed to read environment")
		return
	}
	logger := logrus.NewLogrus(log.NewEntry(log.StandardLogger()))

	// load config
	conf, err := config.Get(e.ConfigPath)
	if err != nil {
		log.WithError(err).Fatal("failed to read config")
		return
	}

	// setup services
	svc := webhook.NewImageWebhook(conf)
	wh, err := mutating.NewWebhook(mutating.WebhookConfig{
		ID:      "image-mutate",
		Mutator: svc,
		Obj:     &corev1.Pod{},
		Logger:  logger,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to setup webhook")
		return
	}
	// create a http handler
	handler, err := whhttp.HandlerFor(whhttp.HandlerConfig{
		Webhook: wh,
		Logger:  logger,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to setup webhook handler")
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
		log.WithField("addr", addr).Info("starting http server")
		log.Fatal(http.ListenAndServeTLS(addr, e.TLS.Cert, e.TLS.Key, router))
	}()

	// wait for a signal
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigC
	log.Printf("received SIGTERM/SIGINT (%s), shutting down...", sig)
}
