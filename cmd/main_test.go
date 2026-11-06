package main

import (
	"testing"

	gardeneroidc "github.com/gardener/oidc-webhook-authenticator/apis/authentication/v1alpha1"
	kyma "github.com/kyma-project/lifecycle-manager/api/v1beta2"
	registrycacheapi "github.com/kyma-project/registry-cache/api/v1beta1"
	"github.com/kyma-project/infrastructure-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestEnableClusterTrustBundleFeatureForSKR(t *testing.T) {
	t.Run("sets all required feature gates and runtime config on empty config", func(t *testing.T) {
		converterConfig := &config.ConverterConfig{}
		enableClusterTrustBundleFeatureForSKR(converterConfig)

		require.NotNil(t, converterConfig.Kubernetes.KubeApiServer.FeatureGates)
		assert.True(t, converterConfig.Kubernetes.KubeApiServer.FeatureGates["ClusterTrustBundle"])
		assert.True(t, converterConfig.Kubernetes.KubeApiServer.FeatureGates["ClusterTrustBundleProjection"])

		require.NotNil(t, converterConfig.Kubernetes.KubeApiServer.RuntimeConfig)
		assert.True(t, converterConfig.Kubernetes.KubeApiServer.RuntimeConfig["certificates.k8s.io/v1beta1/clustertrustbundles"])

		require.NotNil(t, converterConfig.Kubernetes.Kubelet.FeatureGates)
		assert.True(t, converterConfig.Kubernetes.Kubelet.FeatureGates["ClusterTrustBundleProjection"])
		assert.True(t, converterConfig.Kubernetes.Kubelet.FeatureGates["ClusterTrustBundle"])

		require.NotNil(t, converterConfig.Kubernetes.KubeControllerManager.FeatureGates)
		assert.True(t, converterConfig.Kubernetes.KubeControllerManager.FeatureGates["ClusterTrustBundle"])
	})

	t.Run("does not overwrite existing feature gates unrelated to ClusterTrustBundle", func(t *testing.T) {
		converterConfig := &config.ConverterConfig{}
		converterConfig.Kubernetes.KubeApiServer.FeatureGates = map[string]bool{"SomeExistingGate": true}
		converterConfig.Kubernetes.Kubelet.FeatureGates = map[string]bool{"AnotherExistingGate": false}

		enableClusterTrustBundleFeatureForSKR(converterConfig)

		assert.True(t, converterConfig.Kubernetes.KubeApiServer.FeatureGates["SomeExistingGate"])
		assert.True(t, converterConfig.Kubernetes.KubeApiServer.FeatureGates["ClusterTrustBundle"])
		assert.False(t, converterConfig.Kubernetes.Kubelet.FeatureGates["AnotherExistingGate"])
		assert.True(t, converterConfig.Kubernetes.Kubelet.FeatureGates["ClusterTrustBundleProjection"])
	})
}

func TestPrebuiltRuntimeSchemeRegistersTypes(t *testing.T) {
	// Build the scheme the same way main.go does
	prebuiltRuntimeScheme := CreateRuntimeScheme()

	// Verify Kyma Kyma type is registered
	if gvks, _, err := prebuiltRuntimeScheme.ObjectKinds(&kyma.Kyma{}); err != nil {
		t.Fatalf("kyma.Kyma not registered in scheme: %v", err)
	} else if len(gvks) == 0 {
		t.Fatalf("kyma.Kyma returned no GVKs from scheme")
	}

	// Verify registry-cache config type is registered
	if gvks, _, err := prebuiltRuntimeScheme.ObjectKinds(&registrycacheapi.RegistryCacheConfig{}); err != nil {
		t.Fatalf("registry-cache RegistryCacheConfig not registered in scheme: %v", err)
	} else if len(gvks) == 0 {
		t.Fatalf("registry-cache RegistryCacheConfig returned no GVKs from scheme")
	}

	// Verify gardener oidc type is registered
	if gvks, _, err := prebuiltRuntimeScheme.ObjectKinds(&gardeneroidc.OpenIDConnect{}); err != nil {
		t.Fatalf("gardener oidc OpenIDConnect not registered in scheme: %v", err)
	} else if len(gvks) == 0 {
		t.Fatalf("gardener oidc OpenIDConnect returned no GVKs from scheme")
	}

	// Verify apiextensions CRD type is registered
	if gvks, _, err := prebuiltRuntimeScheme.ObjectKinds(&apiextensions.CustomResourceDefinition{}); err != nil {
		t.Fatalf("apiextensions CustomResourceDefinition not registered in scheme: %v", err)
	} else if len(gvks) == 0 {
		t.Fatalf("apiextensions CustomResourceDefinition returned no GVKs from scheme")
	}

	// Verify core v1 Pod is registered (client-go scheme)
	if gvks, _, err := prebuiltRuntimeScheme.ObjectKinds(&corev1.Pod{}); err != nil {
		t.Fatalf("corev1.Pod not registered in scheme: %v", err)
	} else if len(gvks) == 0 {
		t.Fatalf("corev1.Pod returned no GVKs from scheme")
	}
}
