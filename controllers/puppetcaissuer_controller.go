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

	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/camptocamp/puppetca-issuer/provisioners"

	api "github.com/camptocamp/puppetca-issuer/api/v1alpha1"
)

// PuppetCAIssuerReconciler reconciles a PuppetCAIssuer object
type PuppetCAIssuerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=certmanager.puppetca,resources=puppetcaissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certmanager.puppetca,resources=puppetcaissuers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *PuppetCAIssuerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("puppetcaissuer", req.NamespacedName)

	iss := new(api.PuppetCAIssuer)
	if err := r.Client.Get(ctx, req.NamespacedName, iss); err != nil {
		log.Error(err, "failed to retrieve PuppetCAIssuer resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	statusReconciler := newPuppetCAStatusReconciler(r, iss, log)
	if err := validatePuppetCAIssuerSpec(iss.Spec); err != nil {
		log.Error(err, "failed to validate PuppetCAIssuer resource")
		statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "Validation", "Failed to validate resource: %v", err)
		return ctrl.Result{}, err
	}

	// Initialize and store the provisioner

	// Puppet CA url, cert, key, and CA cert are all stored as secrets
	var secret core.Secret
	var ok bool
	var url []byte
	var cert []byte
	var key []byte
	var caCert []byte

	secretNamespaceName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      iss.Spec.Provisioner.Name,
	}

	if err := r.Client.Get(ctx, secretNamespaceName, &secret); err != nil {
		log.Error(err, "failed to retrieve Puppet CA secrets", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		if apierrors.IsNotFound(err) {
			statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "NotFound", "Failed to retrieve Puppet CA secrets: %v", err)
		} else {
			statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "Error", "Failed to retrieve Puppet CA secrets: %v", err)
		}
		return ctrl.Result{}, err
	}

	url, ok = secret.Data[iss.Spec.Provisioner.URLRef.Key]
	if !ok {
		err := fmt.Errorf("secret %s does not contain key %s", secret.Name, iss.Spec.Provisioner.URLRef.Key)
		log.Error(err, "failed to retrieve Puppet CA URL from secret", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "NotFound", "Failed to retrieve Puppet CA URL from secret: %v", err)
		return ctrl.Result{}, err
	}

	cert, ok = secret.Data[iss.Spec.Provisioner.CertRef.Key]
	if !ok {
		err := fmt.Errorf("secret %s does not contain key %s", secret.Name, iss.Spec.Provisioner.CertRef.Key)
		log.Error(err, "failed to retrieve Puppet CA certificate from secret", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "NotFound", "Failed to retrieve Puppet CA certificate from secret: %v", err)
		return ctrl.Result{}, err
	}

	key, ok = secret.Data[iss.Spec.Provisioner.KeyRef.Key]
	if !ok {
		err := fmt.Errorf("secret %s does not contain key %s", secret.Name, iss.Spec.Provisioner.KeyRef.Key)
		log.Error(err, "failed to retrieve Puppet CA key from secret", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "NotFound", "Failed to retrieve Puppet CA key from secret: %v", err)
		return ctrl.Result{}, err
	}

	caCert, ok = secret.Data[iss.Spec.Provisioner.CaCertRef.Key]
	if !ok {
		err := fmt.Errorf("secret %s does not contain key %s", secret.Name, iss.Spec.Provisioner.CaCertRef.Key)
		log.Error(err, "failed to retrieve Puppet CA CA certificate from secret", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		statusReconciler.UpdateNoError(ctx, api.ConditionFalse, "NotFound", "Failed to retrieve Puppet CA CA certificate from secret: %v", err)
		return ctrl.Result{}, err
	}

	p := provisioners.NewProvisioner(string(url), string(cert),
		string(key), string(caCert))

	issNamespaceName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}

	provisioners.Store(issNamespaceName, p)

	return ctrl.Result{}, statusReconciler.Update(ctx, api.ConditionTrue, "Verified", "PuppetCAIssuer verified and ready to sign certificates")
}

func (r *PuppetCAIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.PuppetCAIssuer{}).
		Complete(r)
}

func validatePuppetCAIssuerSpec(s api.PuppetCAIssuerSpec) error {
	switch {
	case s.Provisioner.Name == "":
		return fmt.Errorf("spec.provisioner.name cannot be empty")
	case s.Provisioner.URLRef.Key == "":
		return fmt.Errorf("spec.provisioner.url.key cannot be empty")
	case s.Provisioner.CertRef.Key == "":
		return fmt.Errorf("spec.provisioner.cert.key cannot be empty")
	case s.Provisioner.KeyRef.Key == "":
		return fmt.Errorf("spec.provisioner.key.key cannot be empty")
	case s.Provisioner.CaCertRef.Key == "":
		return fmt.Errorf("spec.provisioner.cacert.key cannot be empty")
	default:
		return nil
	}
}
