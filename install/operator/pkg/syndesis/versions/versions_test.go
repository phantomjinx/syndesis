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

package versions

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/syndesisio/syndesis/install/operator/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/stretchr/testify/assert"

	"github.com/syndesisio/syndesis/install/operator/pkg/apis/syndesis/v1alpha1"
	"github.com/syndesisio/syndesis/install/operator/pkg/apis/syndesis/v1beta1"
	"github.com/syndesisio/syndesis/install/operator/pkg/apis/syndesis/v1beta2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_syndesisAPI_unstructuredToV1Beta2(t *testing.T) {
	type args struct {
		obj unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		wantS   *v1beta2.Syndesis
		wantErr bool
	}{
		{
			"An empty instance v1beta2 should be fine",
			args{obj: getRuntimeObjectAsUnstructured(&v1beta2.Syndesis{})},
			&v1beta2.Syndesis{}, false,
		},
		{
			"An empty instance v1beta1 should be fine",
			args{obj: getRuntimeObjectAsUnstructured(&v1beta1.Syndesis{})},
			&v1beta2.Syndesis{}, false,
		},
		{
			"An empty instance of v1alpha1 should be fine",
			args{obj: getRuntimeObjectAsUnstructured(&v1alpha1.Syndesis{})},
			&v1beta2.Syndesis{}, false,
		},
		{
			"An instance v1beta2 populated should be fine",
			args{obj: getRuntimeObjectAsUnstructured(&v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{Components: v1beta2.ComponentsSpec{Server: v1beta2.ServerConfiguration{Resources: v1beta2.Resources{Limit: v1beta2.ResourceParams{Memory: "800Mi"}}}}}})},
			&v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{Components: v1beta2.ComponentsSpec{Server: v1beta2.ServerConfiguration{Resources: v1beta2.Resources{Limit: v1beta2.ResourceParams{Memory: "800Mi"}}}}}}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := syndesisApi{}
			gotS, err := api.unstructuredToV1Beta2(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("unstructuredToV1Beta2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("unstructuredToV1Beta2() gotS = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}

func Test_syndesisAPI_v1alpha1ToV1beta2(t *testing.T) {
	il := 5
	ici := 10
	dsc := true
	type fields struct {
		v1alpha1 *v1alpha1.Syndesis
		v1beta2  *v1beta2.Syndesis
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantB   *v1beta2.Syndesis
	}{
		{
			"When ForceMigration is true but phase is not Installed, not changes are applied",
			fields{
				v1alpha1: &v1alpha1.Syndesis{Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalling}},
				v1beta2:  &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
		},
		{
			"When phase is Installed but ForceMigration is false, not changes are applied",
			fields{
				v1alpha1: &v1alpha1.Syndesis{Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalled}},
				v1beta2:  &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: false}},
			},
			false,
			&v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: false}},
		},
		{
			"When v1alpha1 is migrated, spec changes",
			fields{
				v1alpha1: &v1alpha1.Syndesis{Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalled}},
				v1beta2:  &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta2.Syndesis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1beta2",
				},
				Spec: v1beta2.SyndesisSpec{ForceMigration: false},
				Status: v1beta2.SyndesisStatus{
					Phase:       v1beta2.SyndesisPhaseInstalled,
					Reason:      v1beta2.SyndesisStatusReasonMigrated,
					Description: fmt.Sprintf("App migrated from %s to %s", v1alpha1.SchemeGroupVersion.String(), v1beta2.SchemeGroupVersion.String()),
				},
			},
		},
		{
			"Migration, check fields",
			fields{
				v1alpha1: &v1alpha1.Syndesis{
					Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalled},
					Spec: v1alpha1.SyndesisSpec{
						SarNamespace:  "sar namespace",
						RouteHostname: "routehostname",
						Addons: v1alpha1.AddonsSpec{
							"ops":    map[string]string{"enabled": "true"},
							"todo":   map[string]string{"enabled": "true"},
							"camelk": map[string]string{"enabled": "false"},
							"jaeger": map[string]string{"enabled": "false"},
						},
						Integration: v1alpha1.IntegrationSpec{
							Limit:              &il,
							StateCheckInterval: &ici,
						},
						MavenRepositories: map[string]string{
							"repo1": "repo1url",
							"repo2": "repo2url",
						},
						Components: v1alpha1.ComponentsSpec{
							Server: v1alpha1.ServerConfiguration{
								Features: v1alpha1.ServerFeatures{
									ManagementUrlFor3scale: "ManagementUrlFor3scale",
								},
								Resources: v1alpha1.Resources{
									ResourceRequirements: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("500m"),
										},
									},
								},
							},
							Db: v1alpha1.DbConfiguration{
								User:     "user",
								Database: "database",
								Resources: v1alpha1.ResourcesWithVolume{
									Resources: v1alpha1.Resources{
										ResourceRequirements: v1.ResourceRequirements{
											Limits: v1.ResourceList{
												v1.ResourceCPU: resource.MustParse("0.5"),
											},
										},
									},
									VolumeCapacity: "1Gi",
								},
							},
							Meta: v1alpha1.MetaConfiguration{
								Resources: v1alpha1.ResourcesWithVolume{
									VolumeCapacity: "5Gi",
									Resources: v1alpha1.Resources{
										ResourceRequirements: v1.ResourceRequirements{
											Limits: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("300m"),
											},
										},
									},
								},
							},
							Prometheus: v1alpha1.PrometheusConfiguration{
								Resources: v1alpha1.ResourcesWithVolume{
									Resources: v1alpha1.Resources{
										ResourceRequirements: v1.ResourceRequirements{
											Limits: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("700m"),
											},
										},
									},
									VolumeCapacity: "2Gi",
								},
							},
							Grafana: v1alpha1.GrafanaConfiguration{
								Resources: v1alpha1.Resources{
									ResourceRequirements: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("500m"),
										},
									},
								},
							},
							Oauth: v1alpha1.OauthConfiguration{
								DisableSarCheck: &dsc,
							},
						},
					},
				},
				v1beta2: &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta2.Syndesis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1beta2",
				},
				Status: v1beta2.SyndesisStatus{
					Phase:       v1beta2.SyndesisPhaseInstalled,
					Reason:      v1beta2.SyndesisStatusReasonMigrated,
					Description: fmt.Sprintf("App migrated from %s to %s", v1alpha1.SchemeGroupVersion.String(), v1beta2.SchemeGroupVersion.String()),
				},
				Spec: v1beta2.SyndesisSpec{
					ForceMigration: false,
					RouteHostname:  "routehostname",
					Addons: v1beta2.AddonsSpec{
						Jaeger: v1beta2.JaegerConfiguration{Enabled: false},
						Ops:    v1beta2.AddonSpec{Enabled: true},
						Todo:   v1beta2.AddonSpec{Enabled: true},
						CamelK: v1beta2.AddonSpec{Enabled: false},
					},
					Components: v1beta2.ComponentsSpec{
						Oauth: v1beta2.OauthConfiguration{SarNamespace: "sar namespace", DisableSarCheck: dsc},
						Server: v1beta2.ServerConfiguration{
							Features: v1beta2.ServerFeatures{
								IntegrationLimit:              il,
								IntegrationStateCheckInterval: ici,
								ManagementUrlFor3scale:        "ManagementUrlFor3scale",
								Maven: v1beta2.MavenConfiguration{
									Repositories: map[string]string{
										"repo1": "repo1url",
										"repo2": "repo2url",
									},
								},
							},
							Resources: v1beta2.Resources{
								Limit: v1beta2.ResourceParams{
									Memory: "500m",
								},
								Request: v1beta2.ResourceParams{
									Memory: "500m",
								},
							},
						},
						Database:   v1beta2.DatabaseConfiguration{User: "user", Name: "database", Resources: v1beta2.ResourcesWithPersistentVolume{VolumeCapacity: "1Gi"}},
						Meta:       v1beta2.MetaConfiguration{Resources: v1beta2.ResourcesWithPersistentVolume{Limit: v1beta2.ResourceParams{Memory: "300m"}, Request: v1beta2.ResourceParams{Memory: "300m"}, VolumeCapacity: "5Gi"}},
						Prometheus: v1beta2.PrometheusConfiguration{Resources: v1beta2.ResourcesWithPersistentVolume{Limit: v1beta2.ResourceParams{Memory: "700m"}, Request: v1beta2.ResourceParams{Memory: "700m"}, VolumeCapacity: "2Gi"}},
						Grafana:    v1beta2.GrafanaConfiguration{Resources: v1beta2.Resources{Limit: v1beta2.ResourceParams{Memory: "500m"}, Request: v1beta2.ResourceParams{Memory: "500m"}}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := syndesisApi{
				v1alpha1: tt.fields.v1alpha1,
				v1beta2:  tt.fields.v1beta2,
			}
			if err := api.v1alpha1ToV1beta2(); (err != nil) != tt.wantErr {
				t.Errorf("v1alpha1ToV1beta2() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.wantB, api.v1beta2)
			}
		})
	}
}

func Test_syndesisAPI_v1beta1ToV1beta2(t *testing.T) {
	il := 5
	ici := 10
	dsc := true
	type fields struct {
		v1beta1 *v1beta1.Syndesis
		v1beta2 *v1beta2.Syndesis
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantB   *v1beta2.Syndesis
	}{
		{
			"When ForceMigration is true but phase is not Installed, not changes are applied",
			fields{
				v1beta1: &v1beta1.Syndesis{Status: v1beta1.SyndesisStatus{Phase: v1beta1.SyndesisPhaseInstalling}},
				v1beta2: &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
		},
		{
			"When phase is Installed but ForceMigration is false, not changes are applied",
			fields{
				v1beta1: &v1beta1.Syndesis{Status: v1beta1.SyndesisStatus{Phase: v1beta1.SyndesisPhaseInstalled}},
				v1beta2: &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: false}},
			},
			false,
			&v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: false}},
		},
		{
			"When v1beta1 is migrated, spec changes",
			fields{
				v1beta1: &v1beta1.Syndesis{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Syndesis",
						APIVersion: "syndesis.io/v1beta1",
					},
					Status: v1beta1.SyndesisStatus{Phase: v1beta1.SyndesisPhaseInstalled}},
				v1beta2: &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta2.Syndesis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1beta2",
				},
				Spec: v1beta2.SyndesisSpec{ForceMigration: false},
				Status: v1beta2.SyndesisStatus{
					Phase:       v1beta2.SyndesisPhaseInstalled,
					Reason:      v1beta2.SyndesisStatusReasonMigrated,
					Description: fmt.Sprintf("App migrated from %s to %s", v1beta1.SchemeGroupVersion.String(), v1beta2.SchemeGroupVersion.String()),
				},
			},
		},
		{
			"Migration, check fields",
			fields{
				v1beta1: &v1beta1.Syndesis{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Syndesis",
						APIVersion: "syndesis.io/v1beta1",
					},
					Status: v1beta1.SyndesisStatus{
						Phase:       v1beta1.SyndesisPhaseInstalled,
						Reason:      v1beta1.SyndesisStatusReasonMigrated,
						Description: fmt.Sprintf("App migrated from %s to %s", v1beta1.SchemeGroupVersion.String(), v1beta1.SchemeGroupVersion.String()),
					},
					Spec: v1beta1.SyndesisSpec{
						ForceMigration: false,
						RouteHostname:  "routehostname",
						Addons: v1beta1.AddonsSpec{
							Jaeger: v1beta1.JaegerConfiguration{Enabled: false},
							Ops:    v1beta1.AddonSpec{Enabled: true},
							Todo:   v1beta1.AddonSpec{Enabled: true},
							CamelK: v1beta1.AddonSpec{Enabled: false},
						},
						Components: v1beta1.ComponentsSpec{
							Oauth: v1beta1.OauthConfiguration{SarNamespace: "sar namespace", DisableSarCheck: dsc},
							Server: v1beta1.ServerConfiguration{
								Features: v1beta1.ServerFeatures{
									IntegrationLimit:              il,
									IntegrationStateCheckInterval: ici,
									ManagementUrlFor3scale:        "ManagementUrlFor3scale",
									Maven: v1beta1.MavenConfiguration{
										Repositories: map[string]string{
											"repo1": "repo1url",
											"repo2": "repo2url",
										},
									},
								},
								Resources: v1beta1.Resources{Memory: "500m"},
							},
							Database:   v1beta1.DatabaseConfiguration{User: "user", Name: "database", Resources: v1beta1.ResourcesWithPersistentVolume{VolumeCapacity: "1Gi"}},
							Meta:       v1beta1.MetaConfiguration{Resources: v1beta1.ResourcesWithPersistentVolume{Memory: "300m", VolumeCapacity: "5Gi"}},
							Prometheus: v1beta1.PrometheusConfiguration{Resources: v1beta1.ResourcesWithPersistentVolume{Memory: "700m", VolumeCapacity: "2Gi"}},
							Grafana:    v1beta1.GrafanaConfiguration{Resources: v1beta1.Resources{Memory: "500m"}},
						},
					},
				},
				v1beta2: &v1beta2.Syndesis{Spec: v1beta2.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta2.Syndesis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1beta2",
				},
				Status: v1beta2.SyndesisStatus{
					Phase:       v1beta2.SyndesisPhaseInstalled,
					Reason:      v1beta2.SyndesisStatusReasonMigrated,
					Description: fmt.Sprintf("App migrated from %s to %s", v1beta1.SchemeGroupVersion.String(), v1beta2.SchemeGroupVersion.String()),
				},
				Spec: v1beta2.SyndesisSpec{
					ForceMigration: false,
					RouteHostname:  "routehostname",
					Addons: v1beta2.AddonsSpec{
						Jaeger: v1beta2.JaegerConfiguration{Enabled: false},
						Ops:    v1beta2.AddonSpec{Enabled: true},
						Todo:   v1beta2.AddonSpec{Enabled: true},
						CamelK: v1beta2.AddonSpec{Enabled: false},
					},
					Components: v1beta2.ComponentsSpec{
						Oauth: v1beta2.OauthConfiguration{SarNamespace: "sar namespace", DisableSarCheck: dsc},
						Server: v1beta2.ServerConfiguration{
							Features: v1beta2.ServerFeatures{
								IntegrationLimit:              il,
								IntegrationStateCheckInterval: ici,
								ManagementUrlFor3scale:        "ManagementUrlFor3scale",
								Maven: v1beta2.MavenConfiguration{
									Repositories: map[string]string{
										"repo1": "repo1url",
										"repo2": "repo2url",
									},
								},
							},
							Resources: v1beta2.Resources{Limit: v1beta2.ResourceParams{Memory: "500m"}, Request: v1beta2.ResourceParams{Memory: "500m"}},
						},
						Database:   v1beta2.DatabaseConfiguration{User: "user", Name: "database", Resources: v1beta2.ResourcesWithPersistentVolume{VolumeCapacity: "1Gi"}},
						Meta:       v1beta2.MetaConfiguration{Resources: v1beta2.ResourcesWithPersistentVolume{Limit: v1beta2.ResourceParams{Memory: "300m"}, Request: v1beta2.ResourceParams{Memory: "300m"}, VolumeCapacity: "5Gi"}},
						Prometheus: v1beta2.PrometheusConfiguration{Resources: v1beta2.ResourcesWithPersistentVolume{Limit: v1beta2.ResourceParams{Memory: "700m"}, Request: v1beta2.ResourceParams{Memory: "700m"}, VolumeCapacity: "2Gi"}},
						Grafana:    v1beta2.GrafanaConfiguration{Resources: v1beta2.Resources{Limit: v1beta2.ResourceParams{Memory: "500m"}, Request: v1beta2.ResourceParams{Memory: "500m"}}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := syndesisApi{
				v1beta1: tt.fields.v1beta1,
				v1beta2: tt.fields.v1beta2,
			}
			if err := api.v1beta1ToV1beta2(); (err != nil) != tt.wantErr {
				t.Errorf("v1beta1ToV1beta2() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.wantB, api.v1beta2)
			}
		})
	}
}

func Test_syndesisApi_v1alpha1ToV1beta1(t *testing.T) {
	il := 5
	ici := 10
	dsc := true
	type fields struct {
		v1alpha1 *v1alpha1.Syndesis
		v1beta1  *v1beta1.Syndesis
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		wantB   *v1beta1.Syndesis
	}{
		{
			"When ForceMigration is true but phase is not Installed, not changes are applied",
			fields{
				v1alpha1: &v1alpha1.Syndesis{Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalling}},
				v1beta1:  &v1beta1.Syndesis{Spec: v1beta1.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta1.Syndesis{Spec: v1beta1.SyndesisSpec{ForceMigration: true}},
		},
		{
			"When phase is Installed but ForceMigration is false, not changes are applied",
			fields{
				v1alpha1: &v1alpha1.Syndesis{Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalled}},
				v1beta1:  &v1beta1.Syndesis{Spec: v1beta1.SyndesisSpec{ForceMigration: false}},
			},
			false,
			&v1beta1.Syndesis{Spec: v1beta1.SyndesisSpec{ForceMigration: false}},
		},
		{
			"When v1alpha1 is migrated, spec changes",
			fields{
				v1alpha1: &v1alpha1.Syndesis{TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1alpha1",
				}, Status: v1alpha1.SyndesisStatus{Phase: v1alpha1.SyndesisPhaseInstalled}},
				v1beta1: &v1beta1.Syndesis{Spec: v1beta1.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta1.Syndesis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1beta1",
				},
				Spec: v1beta1.SyndesisSpec{ForceMigration: false},
				Status: v1beta1.SyndesisStatus{
					Phase:       v1beta1.SyndesisPhaseInstalled,
					Reason:      v1beta1.SyndesisStatusReasonMigrated,
					Description: fmt.Sprintf("App migrated from %s to %s", v1alpha1.SchemeGroupVersion.String(), v1beta1.SchemeGroupVersion.String()),
				},
			},
		},
		{
			"Migration, check fields",
			fields{
				v1alpha1: &v1alpha1.Syndesis{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Syndesis",
						APIVersion: "syndesis.io/v1alpha1",
					},
					// ForceUpgrade is a common field so should be migrated implicitly
					Status: v1alpha1.SyndesisStatus{ForceUpgrade: true, Phase: v1alpha1.SyndesisPhaseInstalled},
					Spec: v1alpha1.SyndesisSpec{
						SarNamespace:  "sar namespace",
						RouteHostname: "routehostname",
						Addons: v1alpha1.AddonsSpec{
							"ops":    map[string]string{"enabled": "true"},
							"todo":   map[string]string{"enabled": "true"},
							"camelk": map[string]string{"enabled": "false"},
							"jaeger": map[string]string{"enabled": "false"},
						},
						Integration: v1alpha1.IntegrationSpec{
							Limit:              &il,
							StateCheckInterval: &ici,
						},
						MavenRepositories: map[string]string{
							"repo1": "repo1url",
							"repo2": "repo2url",
						},
						Components: v1alpha1.ComponentsSpec{
							Server: v1alpha1.ServerConfiguration{
								Features: v1alpha1.ServerFeatures{
									ManagementUrlFor3scale: "ManagementUrlFor3scale",
								},
								Resources: v1alpha1.Resources{
									ResourceRequirements: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("500m"),
										},
									},
								},
							},
							Db: v1alpha1.DbConfiguration{
								User:     "user",
								Database: "database",
								Resources: v1alpha1.ResourcesWithVolume{
									Resources: v1alpha1.Resources{
										ResourceRequirements: v1.ResourceRequirements{
											Limits: v1.ResourceList{
												v1.ResourceCPU: resource.MustParse("0.5"),
											},
										},
									},
									VolumeCapacity: "1Gi",
								},
							},
							Meta: v1alpha1.MetaConfiguration{
								Resources: v1alpha1.ResourcesWithVolume{
									VolumeCapacity: "5Gi",
									Resources: v1alpha1.Resources{
										ResourceRequirements: v1.ResourceRequirements{
											Limits: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("300m"),
											},
										},
									},
								},
							},
							Prometheus: v1alpha1.PrometheusConfiguration{
								Resources: v1alpha1.ResourcesWithVolume{
									Resources: v1alpha1.Resources{
										ResourceRequirements: v1.ResourceRequirements{
											Limits: v1.ResourceList{
												v1.ResourceMemory: resource.MustParse("700m"),
											},
										},
									},
									VolumeCapacity: "2Gi",
								},
							},
							Grafana: v1alpha1.GrafanaConfiguration{
								Resources: v1alpha1.Resources{
									ResourceRequirements: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("500m"),
										},
									},
								},
							},
							Oauth: v1alpha1.OauthConfiguration{
								DisableSarCheck: &dsc,
							},
						},
					},
				},
				v1beta1: &v1beta1.Syndesis{Spec: v1beta1.SyndesisSpec{ForceMigration: true}},
			},
			false,
			&v1beta1.Syndesis{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Syndesis",
					APIVersion: "syndesis.io/v1beta1",
				},
				Status: v1beta1.SyndesisStatus{
					ForceUpgrade: true,
					Phase:        v1beta1.SyndesisPhaseInstalled,
					Reason:       v1beta1.SyndesisStatusReasonMigrated,
					Description:  fmt.Sprintf("App migrated from %s to %s", v1alpha1.SchemeGroupVersion.String(), v1beta1.SchemeGroupVersion.String()),
				},
				Spec: v1beta1.SyndesisSpec{
					ForceMigration: false,
					RouteHostname:  "routehostname",
					Addons: v1beta1.AddonsSpec{
						Jaeger: v1beta1.JaegerConfiguration{Enabled: false},
						Ops:    v1beta1.AddonSpec{Enabled: true},
						Todo:   v1beta1.AddonSpec{Enabled: true},
						CamelK: v1beta1.AddonSpec{Enabled: false},
					},
					Components: v1beta1.ComponentsSpec{
						Oauth: v1beta1.OauthConfiguration{SarNamespace: "sar namespace", DisableSarCheck: dsc},
						Server: v1beta1.ServerConfiguration{
							Features: v1beta1.ServerFeatures{
								IntegrationLimit:              il,
								IntegrationStateCheckInterval: ici,
								ManagementUrlFor3scale:        "ManagementUrlFor3scale",
								Maven: v1beta1.MavenConfiguration{
									Repositories: map[string]string{
										"repo1": "repo1url",
										"repo2": "repo2url",
									},
								},
							},
							Resources: v1beta1.Resources{Memory: "500m"},
						},
						Database:   v1beta1.DatabaseConfiguration{User: "user", Name: "database", Resources: v1beta1.ResourcesWithPersistentVolume{VolumeCapacity: "1Gi"}},
						Meta:       v1beta1.MetaConfiguration{Resources: v1beta1.ResourcesWithPersistentVolume{Memory: "300m", VolumeCapacity: "5Gi"}},
						Prometheus: v1beta1.PrometheusConfiguration{Resources: v1beta1.ResourcesWithPersistentVolume{Memory: "700m", VolumeCapacity: "2Gi"}},
						Grafana:    v1beta1.GrafanaConfiguration{Resources: v1beta1.Resources{Memory: "500m"}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := syndesisApi{
				v1alpha1: tt.fields.v1alpha1,
				v1beta1:  tt.fields.v1beta1,
			}
			if err := api.v1alpha1ToV1beta1(); (err != nil) != tt.wantErr {
				t.Errorf("v1alpha1ToV1beta1() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.wantB, api.v1beta1)
			}
		})
	}
}

func getRuntimeObjectAsUnstructured(obj runtime.Object) (r unstructured.Unstructured) {
	s, err := util.ToUnstructured(obj)
	if err != nil {
		return unstructured.Unstructured{}
	} else {
		return *s
	}
}
