# Secret Manager
Secret Manager is a set of Kubernetes CRDs and controllers which define a common method of interacting with External
SecretStores.

## SecretStore Backends
### Supported
* Hashicorp Vault
* AWS SecretManager
### Planned
* GCP Secret Manager

## Inspiration
Inspired by the great work done by the contributors over at [godaddy/kubernetes-external-secrets][1] and
[jetstack/cert-manager][2], This project aims to take some of the best ideas from both projects for managing secrets.

## Examples
### Basic Example
To use an ExternalSecret first define a SecretStore for use.
```yaml
apiVersion: secret-manager.itscontained.io/v1alpha1
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
        mountPath: kubernetes
        role: example-role
        secretRef:
          name: vault-secret
```

The SecretStore defines how ExternalSecrets for the Store should interact with the backend, and the permission boundary
that the ExternalSecrets have within the namespace or cluster when accessing the SecretStore.

Once a SecretStore is defined an ExternalSecret can be created which references the Store.

In this example, the Vault KV Secrets Engine has a secret at the path `teamA/hello-service`:
```json
{
  "data": {
      "serviceBapiKey": "foo-123",
      "serviceCapiKey": "bar-456",
      "private-images": "{ \"auths\": {\"registry.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"email\":\"foo@example.com\"}}}"
  }
}
```

The ExternalSecret referencing this secret would look like:
```yaml
apiVersion: secret-manager.itscontained.io/v1alpha1
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
      property: serviceBapiKey
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
The ExternalSecret can also optionally define the secret polling time. The ExternalSecret is refreshed when this
period passes.
```yaml
apiVersion: secret-manager.itscontained.io/v1alpha1
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
      property: serviceBapiKey
```

#### Templating Secrets
The ExternalSecret can optionally define the format of the created Kubernetes secrets. The `template` specification
field deeply merges with the generated ExternalSecret and ran through a go template parser. This can allow secrets
with `type` other than `Opaque`, custom labels/annotations on the secret, or a secret data field configured differently
than the data available in the ExternalSecret Store.

An example imagePullSecret with an ExternalSecret:
```yaml
apiVersion: secret-manager.itscontained.io/v1alpha1
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
      property: private-images
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
apiVersion: secret-manager.itscontained.io/v1alpha1
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
      property: serviceBapiKey
  template:
    data:
      config.yaml: |
      {
        "apiUrl": "http://localhost:12345",
        "apiKey": {{ .data.password | quote }}
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
  config.yaml: ewogICJhcGlVcmwiOiAiaHR0cDovL2xvY2FsaG9zdDoxMjM0NSIsCiAgImFwaUtleSI6ICJmb28tMTIzIgp9
# config.yaml: |
# {
#   "apiUrl": "http://localhost:12345"
#   "apiKey": "foo-123"
# }
```


### Embedding Secrets

If the SecretStore returns a map of secret values, then these secrets can be individually referenced via the `property`
field as already demonstrated. When all secret fields should be in the generated secret, the `dataFrom` field can be
specified to fetch all ExternalSecret properties into the generated secret.


```yaml
apiVersion: secret-manager.itscontained.io/v1alpha1
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
  serviceBapiKey: ewogICJjb25maWciOiAiZm9vLTEyMyIKfQ==
  serviceCapiKey: YmFyLTQ1Ng==
  private-images: eyJhdXRocyI6eyJyZWdpc3RyeS5leGFtcGxlLmNvbSI6eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIiwiZW1haWwiOiJmb29AZXhhbXBsZS5jb20ifX19
# "serviceBapiKey": "foo-123",
# "serviceCapiKey": "bar-456",
# "private-images": "{ \"auths\": {\"registry.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"email\":\"foo@example.com\"}}}"
```

[1]: https://github.com/godaddy/kubernetes-external-secrets
[2]: https://github.com/jetstack/cert-manager