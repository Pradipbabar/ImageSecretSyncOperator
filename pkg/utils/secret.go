package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateImagePullSecret creates or updates a docker-registry pull secret in the given namespace
func CreateOrUpdateImagePullSecret(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	secretName, namespace, registry, username, password, email string,
) error {

	auth := fmt.Sprintf("%s:%s", username, password)
	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))

	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			registry: map[string]string{
				"username": username,
				"password": password,
				"email":    email,
				"auth":     authEncoded,
			},
		},
	}

	dockerConfigJSON, err := json.Marshal(dockerConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal docker config: %w", err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": dockerConfigJSON,
		},
	}

	found := &corev1.Secret{}
	err = c.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, found)
	if errors.IsNotFound(err) {
		return c.Create(ctx, secret)
	} else if err != nil {
		return err
	}

	found.Data = secret.Data
	found.Type = corev1.SecretTypeDockerConfigJson
	return c.Update(ctx, found)
}
