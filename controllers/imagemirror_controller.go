/*
Copyright (c) 2020 Facebook, Inc. and its affiliates.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	slipwayk8sfacebookcomv1 "github.com/davidewatson/slipway/api/v1"
)

// ImageMirrorReconciler reconciles a ImageMirror object
type ImageMirrorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=slipway.k8s.facebook.com,resources=imagemirrors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=slipway.k8s.facebook.com,resources=imagemirrors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is called when a resource we are watching may have changed.
func (r *ImageMirrorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var imageMirror slipwayk8sfacebookcomv1.ImageMirror

	ctx := context.Background()
	log := r.Log.WithValues("imagemirror", req.NamespacedName)

	// Get current version of the spec.
	if err := r.Get(ctx, req.NamespacedName, &imageMirror); err != nil {
		log.Error(err, "unable to fetch ImageMirror")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get credentials needed to mirror. We unconditionally read these so that
	// we always have the latest copy, relying on the shared informer cache to
	// avoid unnecessary reads.
	sourceSecretData, err := r.GetSecretData(ctx, imageMirror.ObjectMeta.Namespace, imageMirror.Spec.SourceSecretName)
	if err != nil {
		log.Error(err, "unable to GetSecretData for source")
		return ctrl.Result{}, err
	}
	log.Info("Got source secret", "username", sourceSecretData.Username)

	destSecretData, err := r.GetSecretData(ctx, imageMirror.ObjectMeta.Namespace, imageMirror.Spec.DestSecretName)
	if err != nil {
		log.Error(err, "unable to GetSecretData for dest")
		return ctrl.Result{}, err
	}
	log.Info("Got destination secret", "username", destSecretData.Username)

	// Mirror tags based on the users intent.
	mirroredTags, err := MirrorImages(ctx, log, imageMirror, sourceSecretData, destSecretData)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}
	log.Info("Finished mirroring images")

	// Update status with the current state. Notice that we could have set this
	// within MirrorImages(), but we want crystal clear on when and where state
	// changes.
	imageMirror.Status.MirroredTags = mirroredTags
	if err := r.Status().Update(ctx, &imageMirror); err != nil {
		log.Error(err, "unable to update ImageMirror status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// GetSecretData returns basic credentials from the secret named name in
// namespace, and an err, if any.
func (r *ImageMirrorReconciler) GetSecretData(ctx context.Context, namespace, name string) (data SecretData, err error) {
	if name == "" {
		return data, nil
	}

	// Get the resource using a typed object.
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, secret); err != nil {
		return data, err
	}

	if value, ok := secret.Data["username"]; ok {
		if decoded, err := base64.StdEncoding.DecodeString(string(value)); err == nil {
			data.Username = string(decoded)
		}
	}

	if value, ok := secret.Data["password"]; ok {
		if decoded, err := base64.StdEncoding.DecodeString(string(value)); err != nil {
			data.Password = string(decoded)
		}
	}

	return data, nil
}

// SetupWithManager registers controller with manager and configures shared informer.
func (r *ImageMirrorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&slipwayk8sfacebookcomv1.ImageMirror{}).
		Complete(r)
}
