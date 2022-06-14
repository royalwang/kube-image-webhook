# Kube Image Webhook

The Kube Image [Webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks) allows you to automatically rewrite Pod image references to point to different registries.
This is extremely used in locations without direct internet access (e.g. corporate networks) or if you just want to avoid DockerHub rate limiting.

## Prerequisites

Kubernetes 1.16 or above is required for the `admissionregistration.k8s.io/v1` API.

### Permissions

The application itself doesn't require `cluster-admin`, however the MutatingWebhookConfiguration does due to being a cluster-level resource.

If you are in an environment with restricted permissions, you may separate the MWC and application deployment and ask your administrators to manage the MWC.

### Cert Manager

The webhook expects Cert Manager to be installed so that it can offload creation and management of TLS certificates.
If you are unable to use Cert Manager for any reason, you should manually manage the MutatingWebhookConfiguration.

## Configuration

Configuration is controlled via a YAML file provided to the application.
The format is:

```yaml
images:
  - source: index.docker.io
    destination: docker.example.org
  - source: localhost:5000
    destination: docker.corp.internal:1234
```

### DockerHub

DockerHub has some quirks in that it is the default registry and has special behaviour for "library images".
This webhook handles that by normalising the image before rewriting it.

This means that the `index.docker.io` source will rewrite the following images:
* `ubuntu`
* `bitnami/postgresql`
* `docker.io/ubuntu`

### Chaining

The webhook applies mutation in the order that you define, so you can chain mutations together.
For example:

```yaml
images:
  - source: index.docker.io
    destination: docker.example.org
  - source: docker.example.org/foobar
    destination: some-internal-registry:5000/foobar/zoo
```

If you attempt to pull `foobar/app:latest`, it will end up as `some-internal-registry:5000/foobar/zoo/app:latest`

## Testing

Requires the following:
* Cert Manager
* Development Kubernetes cluster (e.g. `minikube`)
* `kube-image` namespace
* Skaffold
* Kubectl

```bash
skaffold run
```

The Webhook will be deployed to the `kube-image` namespace and will watch the `default` namespace for pods.
The webhook can be tested by creating pods. E.g. 
```
kubectl run -it --rm --image=ubuntu name test -n default
```