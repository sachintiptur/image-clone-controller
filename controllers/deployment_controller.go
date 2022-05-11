/*
Copyright 2022.

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
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	local     = "localhost:65132"
	localSpec = "localhost:5000/"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Deployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var dep appsv1.Deployment
	log := ctrl.LoggerFrom(ctx)
	if err := r.Get(ctx, req.NamespacedName, &dep); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since we can get them on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.V(0).Info("Deployment", "namespace", dep.Namespace)
	if dep.Namespace == "kube-system" {
		log.V(1).Info("Ignore deployments belonging to kube-system namespace")
		return ctrl.Result{}, nil

	}
	imagereg := dep.Spec.Template.Spec.Containers[0].Image
	log.V(1).Info("Image", "registry", imagereg)

	// Ignore if the deployment is already using the backup registry
	if strings.Contains(imagereg, localSpec) {
		log.V(1).Info("Deployment already using local registry")
		return ctrl.Result{}, nil

	}

	ref, err := name.ParseReference(imagereg)
	if err != nil {
		panic(err)
	}

	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		panic(err)
	}
	//log.Println("image", img)

	dstTag := fmt.Sprintf("%s/%s", local, imagereg)
	dstRef, err := name.ParseReference(dstTag)
	if err != nil {
		panic(err)
	}

	//Store the image into local backup registry
	if err := remote.Write(dstRef, img, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		log.Error(err, "Writing error")
	}

	//Update the container image with local registry
	dep.Spec.Template.Spec.Containers[0].Image = localSpec + imagereg
	dep.Spec.Template.Spec.Containers[0].ImagePullPolicy = "Always"

	if err := r.Update(ctx, &dep); err != nil {
		log.Error(err, "unable to update Deployment")
		if apierrors.IsConflict(err) {
			// The Deployment has been updated since we read it.
			// Requeue the Deployment to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		if apierrors.IsNotFound(err) {
			// The Deployment has been deleted since we read it.
			// Requeue the Deployment to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
