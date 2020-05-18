/*
 * Copyright (C) 2019 Red Hat, Inc.
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

package capabilities

import (
	"errors"
	"fmt"

	errs "github.com/pkg/errors"
	"github.com/syndesisio/syndesis/install/operator/pkg/syndesis/clienttools"
)

type ApiServerSpec struct {
	Version          string // Set to the kubernetes version of the API Server
	ImageStreams     bool   // Set to true if the API Server supports imagestreams
	Routes           bool   // Set to true if the API Server supports routes
	EmbeddedProvider bool   // Set to true if the API Server support an embedded authenticaion provider, eg. openshift
	OlmSupport       bool   // Set to true if the API Server supports an Operation-Lifecyle-Manager
}

type RequiredApiSpec struct {
	routes         string
	imagestreams   string
	authprovider   string
	catalogsources string
}

var requiredApi = RequiredApiSpec{
	routes:         "routes.route.openshift.io/v1",
	imagestreams:   "imagestreams.image.openshift.io/v1",
	authprovider:   "oauthclientauthorizations.oauth.openshift.io/v1",
	catalogsources: "catalogsources.operators.coreos.com/v1alpha1",
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

/*
 * For testing if the giving platform is Openshift
 */
func ApiCapabilities(clientTools *clienttools.ClientTools) (*ApiServerSpec, error) {
	apiClient, err := clientTools.ApiClient()
	if err != nil {
		return nil, errs.Wrap(err, "Failed to initialise api client so cannot determine api capabilities")
	}

	if apiClient == nil {
		return nil, errors.New("No api client. Cannot determine api capabilities")
	}

	apiSpec := ApiServerSpec{}

	info, err := apiClient.Discovery().ServerVersion()
	if err != nil {
		return nil, errs.Wrap(err, "Failed to discover server version")
	}

	apiSpec.Version = info.Major + "." + info.Minor

	_, apiResourceLists, err := apiClient.Discovery().ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

	resIndex := []string{}

	for _, apiResList := range apiResourceLists {
		for _, apiResource := range apiResList.APIResources {
			resIndex = append(resIndex, fmt.Sprintf("%s.%s", apiResource.Name, apiResList.GroupVersion))
		}
	}

	apiSpec.Routes = contains(resIndex, requiredApi.routes)
	apiSpec.ImageStreams = contains(resIndex, requiredApi.imagestreams)
	apiSpec.EmbeddedProvider = contains(resIndex, requiredApi.authprovider)
	apiSpec.OlmSupport = contains(resIndex, requiredApi.catalogsources)

	return &apiSpec, nil
}
