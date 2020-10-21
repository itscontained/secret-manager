# secret-manager

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![Version: 0.2.2](https://img.shields.io/badge/Version-0.2.2-informational?style=flat-square) ![AppVersion: %%APP_VERSION%%](https://img.shields.io/badge/AppVersion-%%APP_VERSION%%-informational?style=flat-square)

External secret management for Kubernetes

## TL;DR
```console
helm repo add itscontained https://charts.itscontained.io
helm install itscontained/secret-manager
```

## Installing the Chart
To install the chart with the release name `secret-manager`:
```console
helm install secret-manager itscontained/secret-manager
```

### Custom Resources
By default, the chart will install secret-manager's CRDs, this can be controlled with `installCRDs` value.
> Note: If you're using a Kubernetes version below v1.15 you will need to install the legacy version of the custom resource definitions. The legacy CRDs are maintained on a best-effort basis.

```bash
# Kubernetes <1.15
kubectl apply -f https://raw.githubusercontent.com/itscontained/secret-manager/%%APP_VERSION%%/deploy/crds/legacy/v1beta1_crds_secret-manager.itscontained.io.yaml
```

## Uninstalling the Chart
To uninstall the `secret-manager` deployment:
```console
helm uninstall secret-manager
```
The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration
Read through the secret-manager [values.yaml](https://github.com/itscontained/secret-manager/blob/master/deploy/charts/secret-manager/values.yaml)
file. It has several commented out suggested values.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| apiServerHost | string | `""` | Optional api-server host address to connect to. Only required if out-of-cluster. |
| extraArgs | list | `[]` |  |
| extraEnv | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| healthCheck.enabled | bool | `true` |  |
| healthCheck.port | int | `8400` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"itscontained/secret-manager"` |  |
| image.tag | string | `""` | The image tag to use. The default is the chart appVersion. |
| imagePullSecrets | list | `[]` |  |
| installCRDs | bool | `true` | If set, install and upgrade CRDs through helm chart. |
| kubeConfig | string | `""` | Optionial path to a kubeconfig. Only required if out-of-cluster. |
| leaderElect | bool | `true` | If true, secret-manager will perform leader election between instances to ensure no more than one instance of secret-manager operates at a time. |
| nameOverride | string | `""` |  |
| namespace | string | `""` | If set, this limits the scope of secret-manager to a single namespace and ClusterSecretStores are disabled. If not specified, all namespaces will be watched. |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| prometheus.enabled | bool | `false` |  |
| prometheus.service.annotations | object | `{}` |  |
| prometheus.service.labels | object | `{}` |  |
| prometheus.service.port | int | `9321` |  |
| rbac.create | bool | `true` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created. |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template. |
| tolerations | list | `[]` |  |
