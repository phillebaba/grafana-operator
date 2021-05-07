package grafanaplaylist

import (
	"context"
	"fmt"

	integreatlyv1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/integr8ly/grafana-operator/v3/pkg/controller/common"
)

var log = logf.Log.WithName("controller_grafanaplaylist")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GrafanaDataSource Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, _ chan schema.GroupVersionKind, _ string) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGrafanaPlaylist{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafanaplaylist-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaPlaylist
	err = c.Watch(&source.Kind{Type: &integreatlyv1alpha1.GrafanaPlaylist{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &integreatlyv1alpha1.GrafanaPlaylist{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileGrafanaPlaylist implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGrafanaPlaylist{}

// ReconcileGrafanaPlaylist reconciles a GrafanaPlaylist object
type ReconcileGrafanaPlaylist struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GrafanaPlaylist object and makes changes based on the state read
// and what is in the GrafanaPlaylist.Spec
func (r *ReconcileGrafanaPlaylist) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GrafanaPlaylist")

	// Fetch the GrafanaPlaylist instance
	playlist := &integreatlyv1alpha1.GrafanaPlaylist{}
	err := r.client.Get(context.TODO(), request.NamespacedName, playlist)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Examine if the object is under deletion
	if !playlist.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, *playlist)
	}

	reconciledPlaylist, err := r.reconcile(ctx, *playlist)
	if err != nil {
		return reconcile.Result{}, err
	}

	fmt.Println(reconciledPlaylist)

	return reconcile.Result{}, nil
}

func (r *ReconcileGrafanaPlaylist) reconcile(ctx context.Context, playlist integreatlyv1alpha1.GrafanaPlaylist) (reconcile.Result, error) {

	return reconcile.Result{}, nil
}

func (r *ReconcileGrafanaPlaylist) reconcileDelete(ctx context.Context, playlist integreatlyv1alpha1.GrafanaPlaylist) (reconcile.Result, error) {

	return reconcile.Result{}, nil
}

// Get an authenticated grafana API client
func (r *ReconcileGrafanaPlaylist) getClient() (common.GrafanaClient, error) {
	url := r.state.AdminUrl
	if url == "" {
		return nil, defaultErrors.New("cannot get grafana admin url")
	}

	username := os.Getenv(model.GrafanaAdminUserEnvVar)
	if username == "" {
		return nil, defaultErrors.New("invalid credentials (username)")
	}

	password := os.Getenv(model.GrafanaAdminPasswordEnvVar)
	if password == "" {
		return nil, defaultErrors.New("invalid credentials (password)")
	}

	duration := time.Duration(r.state.ClientTimeout)

	return NewGrafanaClient(url, username, password, r.transport, duration), nil
}
