package webhook

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.com/autokubeops/kube-image-webhook/internal/config"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestImageWebhook_Mutate(t *testing.T) {
	wh := NewImageWebhook(&config.Config{
		Images: []config.Image{
			{
				"index.docker.io",
				"mirror.gcr.io",
			},
		},
	})
	res, err := wh.Mutate(context.TODO(), nil, &corev1.Pod{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name:  "test",
					Image: "ubuntu",
				},
			},
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "busybox",
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, "mirror.gcr.io/library/ubuntu:latest", res.MutatedObject.(*corev1.Pod).Spec.InitContainers[0].Image)
	assert.EqualValues(t, "mirror.gcr.io/library/busybox:latest", res.MutatedObject.(*corev1.Pod).Spec.Containers[0].Image)
}

func TestImageWebhook_normaliseImage(t *testing.T) {
	var cases = []struct {
		in  string
		out string
	}{
		{
			"ubuntu:latest",
			"index.docker.io/library/ubuntu:latest",
		},
		{
			"bitnami/postgresql:latest",
			"index.docker.io/bitnami/postgresql:latest",
		},
		{
			"ubuntu",
			"index.docker.io/library/ubuntu:latest",
		},
		{
			"public.ecr.aws/docker/library/ubuntu:jammy",
			"public.ecr.aws/docker/library/ubuntu:jammy",
		},
		{
			"",
			"",
		},
	}
	wh := &ImageWebhook{}
	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			container := &corev1.Container{
				Image: tt.in,
			}
			wh.normaliseImage(context.TODO(), container)
			assert.EqualValues(t, tt.out, container.Image)
		})
	}
}

func TestImageWebhook_replaceImage(t *testing.T) {
	conf := &config.Config{
		Images: []config.Image{
			{
				Source:      "index.docker.io",
				Destination: "mirror.gcr.io",
			},
			{
				Source:      "public.ecr.aws",
				Destination: "harbor.example.org/public.ecr.aws",
			},
		},
	}
	var cases = []struct {
		in  string
		out string
	}{
		{
			"ubuntu:latest",
			"mirror.gcr.io/library/ubuntu:latest",
		},
		{
			"bitnami/postgresql:latest",
			"mirror.gcr.io/bitnami/postgresql:latest",
		},
		{
			"ubuntu",
			"mirror.gcr.io/library/ubuntu:latest",
		},
		{
			"public.ecr.aws/docker/library/ubuntu:jammy",
			"harbor.example.org/public.ecr.aws/docker/library/ubuntu:jammy",
		},
		{
			"",
			"",
		},
	}
	wh := NewImageWebhook(conf)
	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			container := &corev1.Container{
				Image: tt.in,
			}
			wh.normaliseImage(context.TODO(), container)
			wh.replaceImage(context.TODO(), container)
			assert.EqualValues(t, tt.out, container.Image)
		})
	}
}