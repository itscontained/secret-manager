# Secret Manager
Secret Manager is a set of Kubernetes CRDs and controllers which define a common method of interacting with External
SecretStores.

## SecretStore Backends
### Supported
* Hashicorp Vault
### Planned
* AWS SecretManager
* GCP Secret Manager

## Inspiration
This project is inspired by the great work done by the contributors over at
[godaddy/kubernetes-external-secrets](https://github.com/godaddy/kubernetes-external-secrets) and
[jetstack/cert-manager](https://github.com/jetstack/cert-manager). This project is meant to take some of the best ideas from both
projects for managing secrets.

## Examples
### Basic Example
To use an ExternalSecret first define a SecretStore for use.
```yaml
apiVerson: secret-manager.itscontained.io/v1alpha1
kind: SecretStore
metadata:
  name: vault
  namespace: example-ns
spec:
  vault:
    server: "https://vault.example.com"
    path: secret/data
    auth:
      kubernetes:
        path: kubernetes
        role: example-role
        secretRef:
          name: vault-secret
```

The SecretStore defines how ExternalSecrets for the Store should interact with the backend and the permission boundary
that the ExternalSecrets have within the namespace or cluster when accessing the SecretStore.

Once a SecretStore is defined an ExternalSecret can be created which references the Store.

In this example, the Vault KV Secrets Engine has a secret at the path `teamA/hello-service`:
```json
{
  "data": {
      "frontend": "foo-123",
      "backend": "bar-456",
      "frontend-images": "{ \"auths\": {\"registry.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"email\":\"foo@example.com\"}}}"
  }
}
```

The ExternalSecret referencing this secret would look like:
```yaml
apiVerson: secret-manager.itscontained.io/v1alpha1
kind: ExternalSecret
metadata:
  name: hello-service
  namespace: example-ns
spec:
  storeRef:
    name: vault
  data:
  - secretKey: password
    remoteRef:
      path: teamA/hello-service
      property: frontend
```

This ExternalSecret generates the secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hello-service
  namespace: example-ns
type: Opaque
data:
  # base64 decoded: foo-123
  password: Zm9vLTEyMw==
```

### Advanced Examples
#### Renewing Secrets
The ExternalSecret can also optionally define the secret polling time. The ExternalSecret is refreshed when this period passes.
```yaml
apiVerson: secret-manager.itscontained.io/v1alpha1
kind: ExternalSecret
metadata:
  name: hello-service
  namespace: example-ns
spec:
  storeRef:
    name: vault
  renewAfter: "7d"
  data:
  - secretKey: password
    remoteRef:
      path: teamA/hello-service
      property: frontend
```

#### Templating Secrets
The ExternalSecret can optionally define the format of the created Kubernetes secrets. The `template` specification
field deeply merges with the generated ExternalSecret and ran through a go template parser. This can allow secrets
with `type` other than `Opaque`, custom labels/annotations on the secret, or a secret data field configured differently than the data
available in the ExternalSecret Store.

An example imagePullSecret with an ExternalSecret:
```yaml
apiVerson: secret-manager.itscontained.io/v1alpha1
kind: ExternalSecret
metadata:
  name: hello-service-images
  namespace: example-ns
spec:
  storeRef:
    name: vault
  data:
  - secretKey: .dockerconfigjson
    remoteRef:
      path: teamA/hello-service
      property: frontend-images
  template:
    metadata:
      annotations:
        example: annotation-value
    type: kubernetes.io/dockerconfigjson
```

Generates:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hello-service-images
  namespace: example-ns
  annotations:
    example: annotation-value
type: kubernetes.io/dockerconfigjson
data:
  # base64 decoded: {"auths":{"registry.example.com":{"username":"foo","password":"bar","email":"foo@example.com"}}}
  .dockerconfigjson: eyJhdXRocyI6eyJyZWdpc3RyeS5leGFtcGxlLmNvbSI6eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIiwiZW1haWwiOiJmb29AZXhhbXBsZS5jb20ifX19
```

An example secret with a templated configuration:

```yaml
apiVerson: secret-manager.itscontained.io/v1alpha1
kind: ExternalSecret
metadata:
  name: hello-service-config
  namespace: example-ns
spec:
  storeRef:
    name: vault
  data:
  - secretKey: password
    remoteRef:
      path: teamA/hello-service
      property: frontend
  template:
    password: |
      {
        "config": {{ .Values.data.password | quote }}
      }
```

Generates:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hello-service-config
  namespace: example-ns
type: Opaque
data:
  password: "ewogICJjb25maWciOiAiZm9vLTEyMyIKfQ=="
# password: |
# {
#   "config": "foo-123"
# }
```


### Embedding Secrets

If the SecretStore returns a map of secret values, then these secrets can be individually referenced via the `property` field as already demonstrated. If all secret fields are desired in the generated secret, the `dataFrom` field can be specified to fetch all ExternalSecret properties into the generated secret.


```yaml
apiVerson: secret-manager.itscontained.io/v1alpha1
kind: ExternalSecret
metadata:
  name: hello-service-config
  namespace: example-ns
spec:
  storeRef:
    name: vault
  dataFrom:
  - path: teamA/hello-service
```

Generates:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hello-service-config
  namespace: example-ns
type: Opaque
data:
  frontend: ewogICJjb25maWciOiAiZm9vLTEyMyIKfQ==
  backend: YmFyLTQ1Ng==
  frontend-images: eyJhdXRocyI6eyJyZWdpc3RyeS5leGFtcGxlLmNvbSI6eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIiwiZW1haWwiOiJmb29AZXhhbXBsZS5jb20ifX19
# "frontend": "foo-123",
# "backend": "bar-456",
# "frontend-images": "{ \"auths\": {\"registry.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"email\":\"foo@example.com\"}}}"
```
