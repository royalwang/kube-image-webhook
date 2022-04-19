package webhook

import (
	"context"
	"github.com/google/go-containerregistry/pkg/name"
	log "github.com/sirupsen/logrus"
	"github.com/slok/kubewebhook/v2/pkg/model"
	"github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func NewImageWebhook(c *config.Config) *ImageWebhook {
	return &ImageWebhook{
		conf: c,
	}
}

func (w *ImageWebhook) Mutate(ctx context.Context, _ *model.AdmissionReview, obj metav1.Object) (*mutating.MutatorResult, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return &mutating.MutatorResult{}, nil
	}
	// init containers
	for i := range pod.Spec.InitContainers {
		w.normaliseImage(ctx, &pod.Spec.InitContainers[i])
		w.replaceImage(ctx, &pod.Spec.InitContainers[i])
	}
	// containers
	for i := range pod.Spec.Containers {
		w.normaliseImage(ctx, &pod.Spec.Containers[i])
		w.replaceImage(ctx, &pod.Spec.Containers[i])
	}

	// containers
	return &mutating.MutatorResult{MutatedObject: pod}, nil
}

func (w *ImageWebhook) normaliseImage(ctx context.Context, container *corev1.Container) {
	// if there's no image, there isn't much we can do
	if container.Image == "" {
		log.WithContext(ctx).Warningf("skipping malformed container '%s' - no image present", container.Name)
		return
	}
	ref, err := name.ParseReference(container.Image, name.WithDefaultRegistry("docker.io"))
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to parse image reference")
		return
	}
	log.WithContext(ctx).WithFields(log.Fields{
		"original": ref.String(),
		"current":  ref.Name(),
	}).Info("normalised image reference")
	container.Image = ref.Name()
}

func (w *ImageWebhook) replaceImage(ctx context.Context, container *corev1.Container) {
	// if there's no image, there isn't much we can do
	if container.Image == "" {
		log.WithContext(ctx).Warningf("skipping malformed container '%s' - no image present", container.Name)
		return
	}
	for _, i := range w.conf.Images {
		if strings.HasPrefix(container.Image, i.Source) {
			dst := strings.ReplaceAll(container.Image, i.Source, i.Destination)
			log.WithContext(ctx).WithFields(log.Fields{
				"original": container.Image,
				"current":  dst,
			}).Info("rewrote image reference")
			container.Image = dst
			return
		}
	}
}
