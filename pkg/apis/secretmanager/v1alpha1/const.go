package v1alpha1

import "time"

const (
	DefaultRenewalLeeway = time.Second * 30
	DefaultSecretKey = "secret"

	DefaultVaultAppRoleAuthMountPath    = "approle"
	DefaultVaultKubernetesAuthMountPath = "kubernetes"
)
