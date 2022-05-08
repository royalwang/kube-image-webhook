package webhook

import (
	"github.com/go-logr/logr"
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
)

const DefaultRegistry = "docker.io"

type ImageWebhook struct {
	conf *config.Config
	log  logr.Logger
}
