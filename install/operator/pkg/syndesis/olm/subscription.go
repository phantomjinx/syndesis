/*
 * Copyright (C) 2020 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package olm

import (
	"context"
	"errors"
	"fmt"
	"time"

	olmopv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"

	olmopv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmpkgsvr "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/syndesisio/syndesis/install/operator/pkg/syndesis/clienttools"
	conf "github.com/syndesisio/syndesis/install/operator/pkg/syndesis/configuration"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	pollTimeout  = 600 * time.Second
	pollInterval = 5 * time.Second
)

var sublog = logf.Log.WithName("subscription")

//
// Finds any existing subscriptions for the operator package. If there are none
// then it attempts to create both a subscription and operatorgroup in the
// namespace so that the operator can be initialised.
//
// Returns only those artifacts that are owned by this operator, ie. those
// that have been given ownership by syndesis, allowing these to be tracked
// by the operator and tidied up. If they were not created by this operator
// then they remain independent and will not be tidied up if the CR is removed.
//
func SubscribeOperator(ctx context.Context, clientTools *clienttools.ClientTools, configuration *conf.Config, olmSpec *conf.OlmSpec) error {
	rtClient, err := clientTools.RuntimeClient()
	if err != nil {
		return err
	}

	//
	// Is there Operator-Lifecyle-Manager support?
	//
	if !configuration.ApiServer.OlmSupport {
		return errors.New("Cluster does not support operation-lifecycle-manager")
	}

	sublog.Info("Subscribing to operator", "Package", olmSpec.Package)

	//
	// 1. Look for the packageName in the packageManifest
	//
	pkgManifest, err := findPackageManifest(ctx, rtClient, olmSpec)
	if err != nil {
		return err
	}

	//
	// 2. Check the package has the correct channel
	//
	channel, err := findChannel(ctx, pkgManifest, olmSpec.Channel)

	//
	// 3. Find the CSV supported by the package
	//
	csv, err := findPackageCSV(ctx, rtClient, channel, configuration.OpenShiftProject)
	if err != nil {
		return err
	}

	//
	// 4a. If csv listed with our namespace then an operatorgroup & subscription already installed so RETURN
	//
	if csv != nil {
		//
		// A subscription & operatorgroup exist that will detect our namespace so nothing more to do
		//
		return nil
	}

	//
	// 4b. No csv listed so create the subscription and accompanying operator group
	//
	sub, err := createSubscription(ctx, rtClient, configuration, pkgManifest, channel)
	if err != nil {
		return err
	}

	err = waitForSubscription(ctx, rtClient, sub)
	if err != nil {
		return err
	}

	return nil
}

func findPackageManifest(ctx context.Context, rtClient client.Client, olmSpec *conf.OlmSpec) (*olmpkgsvr.PackageManifest, error) {
	sublog.Info("Finding package manifest for package", "Package", olmSpec.Package)

	//
	// Find the list of package manifests
	//
	pkgs := olmpkgsvr.PackageManifestList{}
	if err := rtClient.List(ctx, &pkgs, &client.ListOptions{Namespace: ""}); err != nil {
		return nil, err
	}

	fmt.Println("Listed package manifests of packages: ", len(pkgs.Items))

	if len(pkgs.Items) == 0 {
		return nil, fmt.Errorf("No package manifests available for Package %s", olmSpec.Package)
	}

	//
	// Find the packagemanifest for the package
	//
	for _, pkg := range pkgs.Items {
		if pkg.Name == olmSpec.Package {
			return &pkg, nil
		}
	}

	return nil, fmt.Errorf("No package manifest available for package %s", olmSpec.Package)
}

func findChannel(ctx context.Context, pkgManifest *olmpkgsvr.PackageManifest, chnlName string) (*olmpkgsvr.PackageChannel, error) {
	for _, channel := range pkgManifest.Status.Channels {
		if channel.Name == chnlName {
			return &channel, nil
		}
	}

	return nil, fmt.Errorf("The package manifest for %s has no channel %s", pkgManifest.Name, chnlName)
}

func findPackageCSV(ctx context.Context, rtClient client.Client, channel *olmpkgsvr.PackageChannel, namespace string) (*olmopv1alpha1.ClusterServiceVersion, error) {
	sublog.Info("Finding csv for package in namespace", "Channel", channel.Name, "Namespace", namespace)

	csv := olmopv1alpha1.ClusterServiceVersion{}
	if err := rtClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: channel.CurrentCSV}, &csv); err != nil {
		if k8serr.IsNotFound(err) {
			return nil, nil // No csvs in namespace
		}

		// A real error occurred
		return nil, err
	}

	return &csv, nil
}

func createSubscription(ctx context.Context, rtClient client.Client, configuration *conf.Config, pkgManifest *olmpkgsvr.PackageManifest, channel *olmpkgsvr.PackageChannel) (*olmopv1alpha1.Subscription, error) {

	ogName := fmt.Sprintf("%s-%s-og", configuration.OpenShiftProject, pkgManifest.Status.PackageName)

	fmt.Println("Creating new subscription for package: ", pkgManifest.Status.PackageName)

	//
	// Create an operator group allowing the OLM to see the namespace
	//
	og := &olmopv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: configuration.OpenShiftProject,
			Name:      ogName,
			Labels:    map[string]string{configuration.ProductName: configuration.OpenShiftProject},
		},
		Spec: olmopv1.OperatorGroupSpec{}, // all namespaces by default
	}

	fmt.Println("Creating a subscription")
	//
	// Create a subscription for the install
	//
	sub := &olmopv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: configuration.OpenShiftProject,
			Name:      pkgManifest.Status.PackageName,
		},
		Spec: &olmopv1alpha1.SubscriptionSpec{
			InstallPlanApproval:    olmopv1alpha1.ApprovalAutomatic,
			Package:                pkgManifest.Status.PackageName,
			CatalogSourceNamespace: pkgManifest.Status.CatalogSourceNamespace,
			CatalogSource:          pkgManifest.Status.CatalogSource,
			Channel:                channel.Name,
		},
	}

	//
	// Add remaining data to the operator group and subscription
	//
	csvDesc := channel.CurrentCSVDesc

	// Add CSV to subscription
	sub.Spec.StartingCSV = channel.CurrentCSV

	// Determine install mode and add target ns to group if install mode does not allow all namespaces
	if !hasInstallMode(csvDesc.InstallModes, olmopv1alpha1.InstallModeTypeAllNamespaces) {
		if hasInstallMode(csvDesc.InstallModes, olmopv1alpha1.InstallModeTypeOwnNamespace) {
			og.Spec.TargetNamespaces = []string{configuration.OpenShiftProject}
		} else if hasInstallMode(csvDesc.InstallModes, olmopv1alpha1.InstallModeTypeSingleNamespace) {
			og.Spec.TargetNamespaces = []string{configuration.OpenShiftProject}
		} else if hasInstallMode(csvDesc.InstallModes, olmopv1alpha1.InstallModeTypeMultiNamespace) {
			og.Spec.TargetNamespaces = []string{configuration.OpenShiftProject}
		}
	}

	err := rtClient.Create(ctx, og)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return nil, err
	}

	err = rtClient.Create(ctx, sub)
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return nil, err
	}

	fmt.Println("Returning subscription")

	return sub, nil
}

func hasInstallMode(installModes []olmopv1alpha1.InstallMode, tgtModeType olmopv1alpha1.InstallModeType) bool {
	if len(installModes) == 0 {
		return false
	}

	for _, installMode := range installModes {
		if installMode.Type == tgtModeType {
			return installMode.Supported
		}
	}

	return false
}

func waitForSubscription(ctx context.Context, rtClient client.Client, sub *olmopv1alpha1.Subscription) error {
	//
	// Wait for the subscription to install the operator
	//
	err := wait.Poll(pollInterval, pollTimeout, func() (done bool, err error) {

		fmt.Println("Polling for install of subscription")
		//
		// Fetch latest information for subscription
		//
		if err := rtClient.Get(ctx, client.ObjectKey{Namespace: sub.Namespace, Name: sub.Name}, sub); err != nil {
			return false, err
		}

		fmt.Println("Checking for install plan reference")
		if sub.Status.InstallPlanRef == nil {
			//
			// No install plan reference so something has gone wrong
			//
			return false, fmt.Errorf("Subscription %s does not have an install plan", sub.Name)
		}

		fmt.Println("Getting install plan for subscription")
		iPlanRef := sub.Status.InstallPlanRef
		installPlan := &olmopv1alpha1.InstallPlan{}
		if err := rtClient.Get(ctx, client.ObjectKey{Namespace: iPlanRef.Namespace, Name: iPlanRef.Name}, installPlan); err != nil {
			return false, err
		}

		fmt.Println("State of install plan: ", installPlan.Status.Phase)

		if installPlan.Status.Phase == olmopv1alpha1.InstallPlanPhaseRequiresApproval {
			return false, fmt.Errorf("Subscription %s requires install approval to complete installation", sub.Name)
		}

		if installPlan.Status.Phase == olmopv1alpha1.InstallPlanPhaseFailed {
			return false, fmt.Errorf("Subscription %s failed to install the operator", sub.Name)
		}

		if installPlan.Status.Phase == olmopv1alpha1.InstallPlanPhaseComplete {
			return true, nil
		}

		fmt.Println("Still waiting on install from install plan of package")
		//
		// Install plan is still to complete so wait
		//
		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}
