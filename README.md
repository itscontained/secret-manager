# Secret Manager

Secret Manager is a set of Kubernetes CRD's and controllers which define a common method of interacting with External SecretStores. Hashicorp's Vault is supported as a SecretStore
with future support planned for AWS SecretManager and GCP Secret Manager.


## Basic Example

Secret Manager takes inspiration from Jetstack's Cert-Manager, there is a seperation of the defined SecretStore (or ClusterSecretStore) and the ExternalSecret. To use an ExternalSecret first define a
SecretStore for use.

```yaml
apiVerson: secret-manager.io/v1alpha1
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

The SecretStore defines how ExternalSecrets for the Store should interact with the backend and the permission boundry the ExternalSecrets have within the namespace or cluster.

Once a SecretStore is defined an ExternalSecret can be created which references the Store.

In this example, the Vault KV Secrets Engine has a secret at `teamA/hello-service`

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
apiVerson: secret-manager.io/v1alpha1
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

And the generated secret would be:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hello-service
  namespace: example-ns
type: Opaque
data:
  password: "Zm9vLTEyMw=="
# password: "foo-123"
```


## Advanced Examples

### Renewing Secrets

The ExternalSecret can also optionally define the period after which the secret should be renewed. The ExternalSecret will be refreshed when this period of time passes.

```yaml
apiVerson: secret-manager.io/v1alpha1
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

### Templating Secrets

The ExternalSecret can optionally define the format of the created Kubernetes secrets. The `template` specification field is deeply merged with the generated ExternalSecret and ran through a go templating parser. This can allow secrets with `type` other than `Opaque`, custom labels/annotations, or a secret data field configured differently than what is availabe in the ExternalSecret Store.

An example imagePullSecret with an ExternalSecret:

```yaml
apiVerson: secret-manager.io/v1alpha1
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
  .dockerconfigjson: "eyJhdXRocyI6eyJyZWdpc3RyeS5leGFtcGxlLmNvbSI6eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIiwiZW1haWwiOiJmb29AZXhhbXBsZS5jb20ifX19"
# .dockerconfigjson: {"auths":{"registry.example.com":{"username":"foo","password":"bar","email":"foo@example.com"}}}

```

An example secret with a templated configuration:

```yaml
apiVerson: secret-manager.io/v1alpha1
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

If the SecretStore can return a map of secrets, these secrets can be individually referenced via the `property` field as already demonstrated. If all fields are desired in the generated secret, the `dataFrom` field can be specified to fetch all ExternalSecret properties into the generated secret.


```yaml
apiVerson: secret-manager.io/v1alpha1
kind: ExternalSecret
metadata:
  name: hello-service-config
  namespace: example-ns
spec:
  storeRef:
    name: vault
  dataFrom:
    remoteRef:
      path: teamA/hello-service
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
  frontend: "ewogICJjb25maWciOiAiZm9vLTEyMyIKfQ=="
  backend: "YmFyLTQ1Ng=="
  frontend-images: "eyJhdXRocyI6eyJyZWdpc3RyeS5leGFtcGxlLmNvbSI6eyJ1c2VybmFtZSI6ImZvbyIsInBhc3N3b3JkIjoiYmFyIiwiZW1haWwiOiJmb29AZXhhbXBsZS5jb20ifX19"
# "frontend": "foo-123",
# "backend": "bar-456",
# "frontend-images": "{ \"auths\": {\"registry.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"email\":\"foo@example.com\"}}}"
```
