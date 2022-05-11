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

	//log "github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"sigs.k8s.io/controller-runtime/pkg/log"
)

// DaemonSetReconciler reconciles a DaemonSet object
type DaemonSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DaemonSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DaemonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Get the daemonset
	var ds appsv1.DaemonSet
	log := ctrl.LoggerFrom(ctx)
	if err := r.Get(ctx, req.NamespacedName, &ds); err != nil {
		log.Error(err, "unable to fetch Daemonset")
		return ctrl.Result{}, err
	}
	log.V(0).Info("Daemonset", "namespace", ds.Namespace)
	if ds.Namespace == "kube-system" {
		log.V(0).Info("Ignore kube-system daemonsets")
		return ctrl.Result{}, nil

	}

	imagereg := ds.Spec.Template.Spec.Containers[0].Image

	ref, err := name.ParseReference(imagereg)
	if err != nil {
		panic(err)
	}

	// Ignore if the daemonset is already using the backup registry
	if strings.Contains(imagereg, localSpec) {
		log.V(0).Info("Daemonset is already using local registry")
		return ctrl.Result{}, nil
	}

	// Read remote image reference
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		panic(err)
	}

	dstTag := fmt.Sprintf("%s/%s", local, imagereg)
	newRef, err := name.ParseReference(dstTag)
	if err != nil {
		panic(err)
	}
	log.V(0).Info("New", "Reference tag", newRef)

	// Store the image into local backup registry
	if err := remote.Write(newRef, img, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		log.Error(err, "Write:")
	}

	//Update the container image
	ds.Spec.Template.Spec.Containers[0].Image = localSpec + imagereg
	ds.Spec.Template.Spec.Containers[0].ImagePullPolicy = "Always"

	if err := r.Update(ctx, &ds); err != nil {
		log.Error(err, "unable to update Deployment")
		if apierrors.IsConflict(err) {
			// The DS has been updated since we read it.
			// Requeue the DS to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		if apierrors.IsNotFound(err) {
			// The DS has been deleted since we read it.
			// Requeue the DS to try to reconciliate again.
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "unable to update Daemonset")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Complete(r)
}
