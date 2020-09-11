/*


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
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	visitorsv1alpha1 "github.com/sourcedelica/visitor-operator/api/v1alpha1"
)

// VisitorsAppReconciler reconciles a VisitorsApp object
type VisitorsAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=visitors.sourcedelica.com,resources=visitorsapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=visitors.sourcedelica.com,resources=visitorsapps/status,verbs=get;update;patch

func (r *VisitorsAppReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("visitorsapp", req.NamespacedName)

	r.Log.Info("Reconciling VisitorsApp")

	// Fetch the VisitorsApp instance
	v := &visitorsv1alpha1.VisitorsApp{}
	err := r.Get(context.TODO(), req.NamespacedName, v)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var result *reconcile.Result

	// == MySQL ==========
	result, err = r.ensureSecret(req, v, r.mysqlAuthSecret(v))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureDeployment(req, v, r.mysqlDeployment(v))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(req, v, r.mysqlService(v))
	if result != nil {
		return *result, err
	}

	mysqlRunning := r.isMysqlUp(v)

	if !mysqlRunning {
		// If MySQL isn't running yet, requeue the reconcile
		// to run again after a delay
		delay := time.Second * time.Duration(5)

		r.Log.Info(fmt.Sprintf("MySQL isn't running, waiting for %s", delay))
		return reconcile.Result{RequeueAfter: delay}, nil
	}

	// == Visitors Backend  ==========
	result, err = r.ensureDeployment(req, v, r.backendDeployment(v))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(req, v, r.backendService(v))
	if result != nil {
		return *result, err
	}

	err = r.updateBackendStatus(v)
	if err != nil {
		// Requeue the request if the status could not be updated
		return reconcile.Result{}, err
	}

	result, err = r.handleBackendChanges(v)
	if result != nil {
		return *result, err
	}

	// == Visitors Frontend ==========
	result, err = r.ensureDeployment(req, v, r.frontendDeployment(v))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(req, v, r.frontendService(v))
	if result != nil {
		return *result, err
	}

	err = r.updateFrontendStatus(v)
	if err != nil {
		// Requeue the request
		return reconcile.Result{}, err
	}

	result, err = r.handleFrontendChanges(v)
	if result != nil {
		return *result, err
	}

	// == Finish ==========
	// Everything went fine, don't requeue
	return reconcile.Result{}, nil
}

func (r *VisitorsAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&visitorsv1alpha1.VisitorsApp{}).
		Complete(r)
}
