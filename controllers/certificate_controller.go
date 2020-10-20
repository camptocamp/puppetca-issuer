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

	api "github.com/camptocamp/puppetca-issuer/api/v1alpha2"
	"github.com/camptocamp/puppetca-issuer/provisioners"
	"github.com/go-logr/logr"
	apiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateReconciler reconciles a PuppetCAIssuer object.
type CertificateReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificate,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificate/status,verbs=get;update;patch

// Reconcile will read and validate a Certificate resource
// and manage the finalizer to delete it on the Puppet CA
func (r *CertificateReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("certificate", req.NamespacedName)

	// Fetch the Certificate resource being reconciled.
	// Just ignore the request if the certificate has been deleted.
	crt := new(cmapi.Certificate)
	if err := r.Client.Get(ctx, req.NamespacedName, crt); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "failed to retrieve Certificate resource")
		return ctrl.Result{}, err
	}

	// Check the Certificate's issuerRef and if it does not match the api
	// group name, log a message at a debug level and stop processing.
	if crt.Spec.IssuerRef.Group != "" && crt.Spec.IssuerRef.Group != api.GroupVersion.Group {
		log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", crt.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}

	// name of our custom finalizer
	myFinalizerName := "puppetca.finalizers.cert-manager.io"

	if crt.ObjectMeta.DeletionTimestamp.IsZero() {
		// Certificate is not being deleted
		if !containsString(crt.ObjectMeta.Finalizers, myFinalizerName) {
			crt.ObjectMeta.Finalizers = append(crt.ObjectMeta.Finalizers, myFinalizerName)
			err := r.Update(context.Background(), crt)
			return ctrl.Result{}, err
		}
	}

	// Certificate is being deleted

	// Do we manage this Certificate?
	if !containsString(crt.ObjectMeta.Finalizers, myFinalizerName) {
		// Not ours to manage
		return ctrl.Result{}, nil
	}

	// Fetch the PuppetCAIssuer resource
	iss := api.PuppetCAIssuer{}
	issNamespaceName := types.NamespacedName{
		Namespace: crt.Namespace,
		Name:      crt.Spec.IssuerRef.Name,
	}
	if err := r.Client.Get(ctx, issNamespaceName, &iss); err != nil {
		log.Error(err, "failed to retrieve PuppetCAIssuer resource", "namespace", crt.Namespace, "name", crt.Spec.IssuerRef.Name)
		_ = r.setStatus(ctx, crt, cmmeta.ConditionFalse, "Pending", "Failed to retrieve PuppetCAIssuer resource %s: %v", issNamespaceName, err)
		return ctrl.Result{}, err
	}

	// Check if the PuppetCAIssuer resource has been marked Ready
	if !PuppetCAIssuerHasCondition(iss, api.PuppetCAIssuerCondition{Type: api.ConditionReady, Status: api.ConditionTrue}) {
		err := fmt.Errorf("resource %s is not ready", issNamespaceName)
		log.Error(err, "failed to retrieve PuppetCAIssuer resource", "namespace", crt.Namespace, "name", crt.Spec.IssuerRef.Name)
		_ = r.setStatus(ctx, crt, cmmeta.ConditionFalse, "Pending", "PuppetCAIssuer resource %s is not Ready", issNamespaceName)
		return ctrl.Result{}, err
	}

	// Load the provisioner that will clean the Certificate
	provisioner, ok := provisioners.Load(issNamespaceName)
	if !ok {
		err := fmt.Errorf("provisioner %s not found", issNamespaceName)
		log.Error(err, "failed to provisioner for PuppetCAIssuer resource")
		_ = r.setStatus(ctx, crt, cmmeta.ConditionFalse, "Pending", "Failed to load provisioner for PuppetCAIssuer resource %s", issNamespaceName)
		return ctrl.Result{}, err
	}

	// Clean Certificate
	if err := provisioner.Clean(ctx, crt); err != nil {
		log.Error(err, "failed to clean certificate")
		return ctrl.Result{}, r.setStatus(ctx, crt, cmmeta.ConditionFalse, "Failed", "Failed to clean certificate: %v", err)
	}

	// Remove finalizer
	crt.ObjectMeta.Finalizers = removeString(crt.ObjectMeta.Finalizers, myFinalizerName)
	if err := r.Update(context.Background(), crt); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.setStatus(ctx, crt, cmmeta.ConditionTrue, "Cleaned", "Certificate cleaned")
}

// SetupWithManager initializes the Certificate controller into the
// controller runtime.
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.Certificate{}).
		Complete(r)
}

func (r *CertificateReconciler) setStatus(ctx context.Context, cr *cmapi.Certificate, status cmmeta.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	apiutil.SetCertificateCondition(cr, cmapi.CertificateConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmeta.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)

	return r.Client.Status().Update(ctx, cr)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
