package controllers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/go-logr/logr"
	corev1alpha1 "github.com/rekuberate-io/sleepcycles/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func (r *SleepCycleReconciler) hasLabel(obj *metav1.ObjectMeta, tag string) bool {
	val, ok := obj.GetLabels()[SleepCycleLabel]

	if ok && val == tag {
		return true
	}

	return false
}

func (r *SleepCycleReconciler) generateToken() (string, error) {
	token := make([]byte, 256)
	_, err := rand.Read(token)
	if err != nil {
		r.logger.Error(err, "error while generating the secret token")
		return "", err
	}

	base64EncodedToken := base64.StdEncoding.EncodeToString(token)
	return base64EncodedToken, nil
}

func (r *SleepCycleReconciler) generateSecureRandomString(length int) (string, error) {
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		return "", err
	}

	for i := range result {
		result[i] = letters[int(result[i])%len(letters)]
	}
	return string(result), nil
}

func (r *SleepCycleReconciler) recordEvent(sleepCycle *corev1alpha1.SleepCycle, message string, isError bool) {
	eventType := corev1.EventTypeNormal
	reason := "SuccessfulSleepCycleReconcile"

	if isError {
		eventType = corev1.EventTypeWarning
		reason = "FailedSleepCycleReconcile"
	}

	r.Recorder.Event(sleepCycle, eventType, reason, strings.ToLower(message))
}

func (r *SleepCycleReconciler) checkCrdExists(ctx context.Context, logger logr.Logger, customResourceDefinitionName string) (bool, error) {
	qsCrd := &extv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind: "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "fleetautoscalers.autoscaling.agones.dev",
		},
	}

	logger.Info("Read the ConsoleQuickStart CRD")
	if err := r.Get(ctx, client.ObjectKeyFromObject(qsCrd), qsCrd); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
