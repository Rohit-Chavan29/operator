/*
Copyright 2025.

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

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mygroupv1 "github.com/Rohit-Chavan29/operator/api/v1"
)

// MyResourceReconciler reconciles a MyResource object
type MyResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mygroup.example.com,resources=myresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mygroup.example.com,resources=myresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mygroup.example.com,resources=myresources/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
    var myResource mygroupv1.MyResource
	if err := r.Get(ctx, req.NamespacedName, &myResource); err != nil {
		log.Error(err, "Failed to fetch MyResource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Define the desired Deployment name
	deploymentName := myResource.Name + "-deployment"

	// Check if Deployment already exists
	var deployment appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: myResource.Namespace}, &deployment)
	if err == nil {
		log.Info("Deployment already exists, skipping creation", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		return ctrl.Result{}, nil
	}

	// Define a new Deployment
	newDeployment := r.createDeployment(&myResource)

	// Create the Deployment
	if err := r.Create(ctx, newDeployment); err != nil {
		log.Error(err, "Failed to create Deployment")
		return ctrl.Result{}, err
	}

	log.Info("Deployment created successfully", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)

	return ctrl.Result{}, nil
}
func (r *MyResourceReconciler) createDeployment(myResource *mygroupv1.MyResource) *appsv1.Deployment {
	labels := map[string]string{"app": myResource.Name}
	replicas := int32(myResource.Spec.Replicas)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myResource.Name + "-deployment",
			Namespace: myResource.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
		
			Replicas: &replicas, // Use replica count from CRD
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-app",
							Image: myResource.Spec.Image, // Use image from CRD
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
	For(&mygroupv1.MyResource{}).
	Owns(&appsv1.Deployment{}).
	Complete(r)
}
