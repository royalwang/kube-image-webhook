package webhook

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/slok/kubewebhook/v2/pkg/model"
	"github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func NewImageWebhook(log logr.Logger, c *config.Config) *ImageWebhook {
	return &ImageWebhook{
		log:  log,
		conf: c,
	}
}

func (w *ImageWebhook) Mutate(_ context.Context, _ *model.AdmissionReview, obj metav1.Object) (*mutating.MutatorResult, error) {
	log := w.log
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		log.V(1).Info("skipping mutation as the type could not be cast to corev1.Pod")
		return &mutating.MutatorResult{}, nil
	}
	// init containers
	log.V(1).Info("mutating initContainers")
	for i := range pod.Spec.InitContainers {
		w.normaliseImage(&pod.Spec.InitContainers[i])
		w.replaceImage(&pod.Spec.InitContainers[i])
	}
	// containers
	log.V(1).Info("mutating containers")
	for i := range pod.Spec.Containers {
		w.normaliseImage(&pod.Spec.Containers[i])
		w.replaceImage(&pod.Spec.Containers[i])
	}

	// containers
	return &mutating.MutatorResult{MutatedObject: pod}, nil
}

func (w *ImageWebhook) normaliseImage(container *corev1.Container) {
	log := w.log.WithValues("Name", container.Name, "Image", container.Image)
	log.V(1).Info("normalising container")
	log.V(3).Info("received container", "Container", container)
	// if there's no image, there isn't much we can do
	if container.Image == "" {
		log.Info("skipping malformed container - no image present")
		return
	}
	log.V(1).Info("parsing image reference")
	ref, err := name.ParseReference(container.Image, name.WithDefaultRegistry(DefaultRegistry))
	if err != nil {
		log.Error(err, "failed to parse image reference")
		return
	}
	log.Info("normalised image reference", "Original", ref.String(), "Current", ref.Name())
	container.Image = ref.Name()
}

func (w *ImageWebhook) replaceImage(container *corev1.Container) {
	log := w.log.WithValues("Name", container.Name, "Image", container.Image)
	log.V(1).Info("rewriting container image")
	log.V(3).Info("received container", "Container", container)
	// if there's no image, there isn't much we can do
	if container.Image == "" {
		log.Info("skipping malformed container - no image present")
		return
	}
	for _, i := range w.conf.Images {
		log.V(2).Info("matching source", "Source", i.Source)
		if strings.HasPrefix(container.Image, i.Source) {
			dst := strings.ReplaceAll(container.Image, i.Source, i.Destination)
			log.Info("rewrote image reference", "Original", container.Image, "Current", dst)
			container.Image = dst
		}
	}
}
