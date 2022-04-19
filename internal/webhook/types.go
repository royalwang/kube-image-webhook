package webhook

import (
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
)

const DefaultRegistry = "docker.io"

type ImageWebhook struct {
	conf *config.Config
}
