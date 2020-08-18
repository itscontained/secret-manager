package v1alpha1

import "time"

const (
	DefaultRenewalLeeway = time.Second * 30

	DefaultVaultAppRoleAuthMountPath    = "approle"
	DefaultVaultKubernetesAuthMountPath = "kubernetes"
)
