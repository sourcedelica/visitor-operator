package controllers

import (
	"context"

	visitorsv1alpha1 "github.com/sourcedelica/visitor-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *VisitorsAppReconciler) ensureDeployment(request reconcile.Request,
	instance *visitorsv1alpha1.VisitorsApp,
	dep *appsv1.Deployment,
) (*reconcile.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment
		r.Log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(context.TODO(), dep)

		if err != nil {
			// Deployment failed
			r.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		r.Log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *VisitorsAppReconciler) ensureService(request reconcile.Request,
	instance *visitorsv1alpha1.VisitorsApp,
	s *corev1.Service,
) (*reconcile.Result, error) {
	found := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		r.Log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			r.Log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		r.Log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func (r *VisitorsAppReconciler) ensureSecret(request reconcile.Request,
	instance *visitorsv1alpha1.VisitorsApp,
	s *corev1.Secret,
) (*reconcile.Result, error) {
	found := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		r.Log.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			r.Log.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return &reconcile.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		r.Log.Error(err, "Failed to get Secret")
		return &reconcile.Result{}, err
	}

	return nil, nil
}

func labels(v *visitorsv1alpha1.VisitorsApp, tier string) map[string]string {
	return map[string]string{
		"app":             "visitors",
		"visitorssite_cr": v.Name,
		"tier":            tier,
	}
}
