package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	imageregistrycredentialv1alpha1 "github.com/Pradipbabar/ImageSecretSyncOperator/api/v1alpha1"
	"github.com/Pradipbabar/ImageSecretSyncOperator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClustRegCredReconciler reconciles a ClustRegCred object
type ClustRegCredReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=imageregistrycredential.pradix.io,resources=clustregcreds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=imageregistrycredential.pradix.io,resources=clustregcreds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=imageregistrycredential.pradix.io,resources=clustregcreds/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;create;update;patch;delete

func (r *ClustRegCredReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cred imageregistrycredentialv1alpha1.ClustRegCred
	if err := r.Get(ctx, req.NamespacedName, &cred); err != nil {
		logger.Error(err, "Unable to fetch ClustRegCred")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Base64 decode the password
	decodedPassword, err := base64.StdEncoding.DecodeString(cred.Spec.Password)
	if err != nil {
		msg := "Failed to decode base64-encoded password"
		logger.Error(err, "Failed to decode base64-encoded password in ClustRegCred")
		r.Recorder.Event(&cred, "Warning", "DecodeFailed", "Failed to decode password")
		setCondition(&cred, "Ready", metav1.ConditionFalse, "DecodeError", msg)
		return ctrl.Result{}, err
	}

	var syncedNamespaces []string

	for _, namespace := range cred.Spec.Namespaces {
		err := utils.CreateOrUpdateImagePullSecret(
			ctx,
			r.Client,
			r.Scheme,
			cred.Spec.SecretName,
			namespace,
			cred.Spec.Registry,
			cred.Spec.Username,
			string(decodedPassword),
			cred.Spec.Email,
		)
		if err != nil {
			msg := fmt.Sprintf("Failed to sync secret in namespace %s: %v", namespace, err)
			logger.Error(err, "Failed to create/update secret", "namespace", namespace)
			r.Recorder.Eventf(&cred, "Warning", "SecretSyncFailed", "Failed to sync secret to namespace %s: %v", namespace, err)

			// Update status on failure
			cred.Status.Phase = "Failed"
			setCondition(&cred, "Ready", metav1.ConditionFalse, "SyncError", msg)
			cred.Status.Reason = err.Error()
			r.Status().Update(ctx, &cred)

			return ctrl.Result{}, err
		}
		syncedNamespaces = append(syncedNamespaces, namespace)
		r.Recorder.Eventf(&cred, "Normal", "SecretSynced", "Synced image pull secret to namespace %s", namespace)
	}

	// Update status after successful sync
	cred.Status.Phase = "Synced"
	cred.Status.SyncedNamespaces = syncedNamespaces
	cred.Status.LastSynced = time.Now().Format(time.RFC3339)
	cred.Status.Reason = "Secrets synced successfully"
	setCondition(&cred, "Ready", metav1.ConditionTrue, "SecretsSynced", "Secrets successfully synced to all namespaces")
	if err := r.Status().Update(ctx, &cred); err != nil {
		logger.Error(err, "Failed to update ClustRegCred status")
		return ctrl.Result{}, err
	}

	r.Recorder.Event(&cred, "Normal", "SyncComplete", "All secrets successfully synced")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClustRegCredReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the event recorder here
	r.Recorder = mgr.GetEventRecorderFor("clustregcred-controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&imageregistrycredentialv1alpha1.ClustRegCred{}).
		Complete(r)
}

func setCondition(cred *imageregistrycredentialv1alpha1.ClustRegCred, condType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&cred.Status.Conditions, metav1.Condition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
}
