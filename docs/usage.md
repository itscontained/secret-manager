# Usage

## Basic example

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
    path: secret
    auth:
      kubernetes:
        mountPath: kubernetes
        role: example-role
        secretRef:
          name: vault-secret

```

The SecretStore defines how ExternalSecrets for the Store should interact with the backend, and the permission boundary that the ExternalSecrets have within the namespace or cluster when accessing the SecretStore.

Once a SecretStore is defined an ExternalSecret can be created which references the Store.

In this example, the Vault KV Secrets Engine has a secret at the name teamA/hello-service:

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
      name: teamA/hello-service
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

## Embedding Secrets

If the SecretStore returns a map of secret values, then these secrets can be individually referenced via the property field as already demonstrated. When all secret fields should be in the generated secret, the dataFrom field can be specified to fetch all ExternalSecret properties into the generated secret.

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
  - name: teamA/hello-service
```

Generates:

```
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
