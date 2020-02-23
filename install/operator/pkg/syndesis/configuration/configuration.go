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

package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/imdario/mergo"

	"k8s.io/apimachinery/pkg/types"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/syndesisio/syndesis/install/operator/pkg/apis/syndesis/v1beta1"
	"github.com/syndesisio/syndesis/install/operator/pkg/util"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

// Location from where the template configuration is located
var TemplateConfig string

type Config struct {
	AllowLocalHost             bool
	Productized                bool
	DevSupport                 bool     // If set to true, pull docker images from imagetag instead of upstream source
	Scheduled                  bool     // Legacy parameter to set scheduled:true in the imagestreams, but we dont use many imagestreams nowadays
	ProductName                string   // Usually syndesis or fuse-online
	PrometheusRules            string   // If some extra rules for prometheus need to be specified, they are defined here
	OpenShiftProject           string   // The name of the OpenShift project Syndesis is being deployed into
	OpenShiftOauthClientSecret string   // OpenShift OAuth client secret
	OpenShiftConsoleUrl        string   // The URL to the OpenShift console
	ImagePullSecrets           []string // Pull secrets attached to services accounts. This field is generated by the operator
	DatabaseNeedsUpgrade       bool           // Enabled the image running the database doesn't match the operator's configured image spec
	Syndesis                   SyndesisConfig // Configuration for syndesis components and addons. This fields are overwritten from environment variables and from the custom resource
}

type SyndesisConfig struct {
	DemoData      bool           // Enables starting up with demo data
	RouteHostname string         // The external hostname to access Syndesis
	Components    ComponentsSpec // Server, Meta, Ui, Name specifications and configurations
	Addons        AddonsSpec     // Addons specifications and configurations
}

// Components
type ComponentsSpec struct {
	UI         UIConfiguration         // Configuration ui
	S2I        S2IConfiguration        // Configuration s2i
	Oauth      OauthConfiguration      // Configuration oauth
	Server     ServerConfiguration     // Configuration server
	Meta       MetaConfiguration       // Configuration meta
	Database   DatabaseConfiguration   // Configuration database
	Prometheus PrometheusConfiguration // Configuration monitoring
	Grafana    GrafanaConfiguration    // Configuration grafana
	Upgrade    UpgradeConfiguration    // Configuration upgrade
}

type OauthConfiguration struct {
	CookieSecret    string // Secret to use to encrypt oauth cookies
	Image           string // Docker image for Oauth
	DisableSarCheck bool   // Enable or disable SAR checks all together
	SarNamespace    string // The user needs to have permissions to at least get a list of pods in the given project in order to be granted access to the Syndesis installation
}

type UIConfiguration struct {
	Image string // Docker image for ui pod
}

type S2IConfiguration struct {
	Image string // Docker image for s2i pod
}

type DatabaseConfiguration struct {
	User             string                        // Username for PostgreSQL user that will be used for accessing the database
	Name             string                        // Name of the PostgreSQL database accessed
	URL              string                        // Host and port of the PostgreSQL database to access
	ExternalDbURL    string                        // If specified, use an external database instead of the installed by syndesis
	Resources        ResourcesWithPersistentVolume // Resources, memory and database volume size
	Exporter         ExporterConfiguration         // The exporter exports metrics in prometheus format
	Image            string                        // Docker image for database
	Password         string                        // Password for the PostgreSQL connection user
	SampledbPassword string                        // Password for the PostgreSQL sampledb user
}

type ExporterConfiguration struct {
	Image string // Docker image for postgres_exporter
}

type PrometheusConfiguration struct {
	Image     string              // Docker image for prometheus
	Rules     string              // Monitoring rules for prometheus
	Resources ResourcesWithVolume // Set volume size for prometheus pod, where metrics are stored
}

type GrafanaConfiguration struct {
	Resources Resources // Resources for grafana pod, memory
}

type ServerConfiguration struct {
	Resources                    Resources      // Resources reserved for server pod
	Features                     ServerFeatures // Server features: integration limits and check interval, support for demo data and more
	Image                        string         // Docker image for server
	SyndesisEncryptKey           string         // The encryption key used to encrypt/decrypt stored secrets
	ClientStateAuthenticationKey string         // Key used to perform authentication of client side stored state
	ClientStateEncryptionKey     string         // Key used to perform encryption of client side stored state
}

type MetaConfiguration struct {
	Image     string              // Docker image for meta
	Resources ResourcesWithVolume // Resources for meta pod, memory
}

type UpgradeConfiguration struct {
	Image     string              // Docker image for Upgrade pod
	Resources VolumeOnlyResources // Resources for upgrade pod, memory and volume size where database dump is saved
}

type Resources struct {
	Memory string
}

type ResourcesWithVolume struct {
	Memory         string
	VolumeCapacity string
}

type ResourcesWithPersistentVolume struct {
	Memory             string
	VolumeCapacity     string
	VolumeName         string
	VolumeAccessMode   string
	VolumeStorageClass string
	VolumeLabels       map[string]string
}

type VolumeOnlyResources struct {
	VolumeCapacity string
}

type ServerFeatures struct {
	IntegrationLimit              int               // Maximum number of integrations single user can create
	IntegrationStateCheckInterval int               // Interval for checking the state of the integrations
	DeployIntegrations            bool              // Whether we deploy integrations
	TestSupport                   bool              // Enables test-support endpoint on backend API
	OpenShiftMaster               string            // Public OpenShift master address
	ManagementUrlFor3scale        string            // 3scale management URL
	MavenRepositories             map[string]string // Set repositories for maven
}

// Addons
type AddonsSpec struct {
	Jaeger    JaegerConfiguration
	Ops       AddonConfiguration
	Todo      AddonConfiguration
	Knative   AddonConfiguration
	DV        DvConfiguration
	CamelK    CamelKConfiguration
	PublicApi PublicApiConfiguration
}

type JaegerConfiguration struct {
	Enabled       bool
	ClientOnly    bool
	OperatorOnly  bool
	QueryUri      string
	CollectorUri  string
	SamplerType   string
	SamplerParam  string
	ImageAgent    string
	ImageAllInOne string
	ImageOperator string
}

type DvConfiguration struct {
	Enabled   bool
	Resources Resources
	Image     string
}

type PublicApiConfiguration struct {
	Enabled         bool
	RouteHostname   string
	DisableSarCheck bool
}

type AddonConfiguration struct {
	Enabled bool
}

type CamelKConfiguration struct {
	Enabled       bool
	CamelVersion  string
	CamelKRuntime string
	Image         string
}

type AddonInstance struct {
	Name    string
	Enabled bool
}

const (
	SyndesisGlobalConfigSecret = "syndesis-global-config"
)

/*
/ Returns an array of the addons names and if configuration has been defined
/ whether they've been enabled in that configuration instance
*/
func GetAddons(configuration Config) []AddonInstance {
	return []AddonInstance{
		{"jaeger", configuration.Syndesis.Addons.Jaeger.Enabled},
		{"ops", configuration.Syndesis.Addons.Ops.Enabled},
		{"dv", configuration.Syndesis.Addons.DV.Enabled},
		{"camelk", configuration.Syndesis.Addons.CamelK.Enabled},
		{"knative", configuration.Syndesis.Addons.Knative.Enabled},
		{"publicApi", configuration.Syndesis.Addons.PublicApi.Enabled},
		{"todo", configuration.Syndesis.Addons.Todo.Enabled},
	}
}

/*
/ Returns all processed configurations for Syndesis

 - Default values for configuration are loaded from file
 - Secrets and passwords are loaded from syndesis-global-config Secret if they exits
 and generated if they dont
 - For QE, some fields are loaded from environment variables
 - Users might define fields using the syndesis custom resource
*/
func GetProperties(file string, ctx context.Context, client client.Client, syndesis *v1beta1.Syndesis) (*Config, error) {
	configuration := &Config{}
	if err := configuration.loadFromFile(file); err != nil {
		return nil, err
	}

	configuration.OpenShiftProject = syndesis.Namespace
	configuration.Syndesis.Components.Oauth.SarNamespace = configuration.OpenShiftProject

	if client != nil {
		if err := configuration.setPasswordsFromSecret(ctx, client, syndesis); err != nil {
			return nil, err
		}
	}
	configuration.generatePasswords()

	if err := configuration.setConfigFromEnv(); err != nil {
		return nil, err
	}

	if err := configuration.setSyndesisFromCustomResource(syndesis); err != nil {
		return nil, err
	}

	if client != nil {
		//
		// If the image running the db container no longer matches the image specified by
		// the operator configuration then the database needs to be upgraded using the internal
		// postgresql upgrade procedure. This will flag that process to be initiated.
		//
		databaseDeployment := &appsv1.DeploymentConfig{}
		if err := client.Get(ctx, types.NamespacedName{Namespace: syndesis.Namespace, Name: "syndesis-db"}, databaseDeployment); err == nil {
			for _, c := range databaseDeployment.Spec.Template.Spec.Containers {
				if c.Name == "postgresql" {
					configuration.DatabaseNeedsUpgrade = c.Image != configuration.Syndesis.Components.Database.Image
					break
				}
			}
		}
	}

	return configuration, nil
}

// Load configuration from config file. Config file is expected to be a yaml
// The returned configuration is parsed to JSON and returned as a Config object
func (config *Config) loadFromFile(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
		data, err = yaml.ToJSON(data)
		if err != nil {
			return err
		}
	}

	if err := json.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

// Set Config.RouteHostname based on the Spec.Host property of the syndesis route
// If an environment variable is set to overwrite the route, take that instead
func (config *Config) SetRoute(ctx context.Context, client client.Client, syndesis *v1beta1.Syndesis) error {
	if os.Getenv("ROUTE_HOSTNAME") == "" {
		syndesisRoute := &routev1.Route{}

		if err := client.Get(ctx, types.NamespacedName{Namespace: syndesis.Namespace, Name: "syndesis"}, syndesisRoute); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			} else {
				return err
			}
		}
		config.Syndesis.RouteHostname = syndesisRoute.Spec.Host
	} else {
		config.Syndesis.RouteHostname = os.Getenv("ROUTE_HOSTNAME")
	}
	return nil
}

// When an external database is defined, reset connection parameters
func (config *Config) ExternalDatabase(ctx context.Context, client client.Client, syndesis *v1beta1.Syndesis) error {
	// Handle an external database being defined
	if syndesis.Spec.Components.Database.ExternalDbURL != "" {

		// setup connection string from provided url
		externalDbURL, err := url.Parse(syndesis.Spec.Components.Database.ExternalDbURL)
		if err != nil {
			return err
		}
		if externalDbURL.Path == "" {
			externalDbURL.Path = syndesis.Spec.Components.Database.Name
		}

		config.Syndesis.Components.Database.URL = externalDbURL.String()
	}

	return nil
}

func getSyndesisConfigurationSecret(ctx context.Context, client client.Client, namespace string) (*corev1.Secret, error) {
	secret := corev1.Secret{}
	if err := client.Get(ctx, util.NewObjectKey(SyndesisGlobalConfigSecret, namespace), &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

func (config *Config) setPasswordsFromSecret(ctx context.Context, client client.Client, syndesis *v1beta1.Syndesis) error {
	secret, err := getSyndesisConfigurationSecret(ctx, client, syndesis.Namespace)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}

	/*
	 * If none exist in the secret then config property is set to ""
	 * If this is the case then passwords are generated as a result of
	 * the call to generatePasswords() following execution of this function
	 */
	if _, ok := secret.Data["POSTGRESQL_PASSWORD"]; !ok {
		// This is an indicator that the secret has the old format. We need to extract the
		// secrets from the `params` section instead
		// TODO: Delete for 1.10
		envFromSecret, err := getSyndesisEnvVarsFromOpenShiftNamespace(secret)
		if err != nil {
			return err
		}

		config.OpenShiftOauthClientSecret = envFromSecret["OPENSHIFT_OAUTH_CLIENT_SECRET"]
		config.Syndesis.Components.Database.Password = envFromSecret["POSTGRESQL_PASSWORD"]
		config.Syndesis.Components.Database.SampledbPassword = envFromSecret["POSTGRESQL_SAMPLEDB_PASSWORD"]
		config.Syndesis.Components.Oauth.CookieSecret = envFromSecret["OAUTH_COOKIE_SECRET"]
		config.Syndesis.Components.Server.SyndesisEncryptKey = envFromSecret["SYNDESIS_ENCRYPT_KEY"]
		config.Syndesis.Components.Server.ClientStateAuthenticationKey = envFromSecret["CLIENT_STATE_AUTHENTICATION_KEY"]
		config.Syndesis.Components.Server.ClientStateEncryptionKey = envFromSecret["CLIENT_STATE_ENCRYPTION_KEY"]
	} else {
		// This is the behaviour we want
		config.OpenShiftOauthClientSecret = string(secret.Data["OPENSHIFT_OAUTH_CLIENT_SECRET"])
		config.Syndesis.Components.Database.Password = string(secret.Data["POSTGRESQL_PASSWORD"])
		config.Syndesis.Components.Database.SampledbPassword = string(secret.Data["POSTGRESQL_SAMPLEDB_PASSWORD"])
		config.Syndesis.Components.Oauth.CookieSecret = string(secret.Data["OAUTH_COOKIE_SECRET"])
		config.Syndesis.Components.Server.SyndesisEncryptKey = string(secret.Data["SYNDESIS_ENCRYPT_KEY"])
		config.Syndesis.Components.Server.ClientStateAuthenticationKey = string(secret.Data["CLIENT_STATE_AUTHENTICATION_KEY"])
		config.Syndesis.Components.Server.ClientStateEncryptionKey = string(secret.Data["CLIENT_STATE_ENCRYPTION_KEY"])
	}

	return nil
}

// Overwrite operand images with values from ENV if those env are present
func (config *Config) setConfigFromEnv() error {
	imgEnv := Config{
		Syndesis: SyndesisConfig{
			Addons: AddonsSpec{
				DV:     DvConfiguration{Image: os.Getenv("DV_IMAGE")},
				CamelK: CamelKConfiguration{Image: os.Getenv("CAMELK_IMAGE")},
			},
			Components: ComponentsSpec{
				Oauth:      OauthConfiguration{Image: os.Getenv("OAUTH_IMAGE")},
				UI:         UIConfiguration{Image: os.Getenv("UI_IMAGE")},
				S2I:        S2IConfiguration{Image: os.Getenv("S2I_IMAGE")},
				Prometheus: PrometheusConfiguration{Image: os.Getenv("PROMETHEUS_IMAGE")},
				Upgrade:    UpgradeConfiguration{Image: os.Getenv("UPGRADE_IMAGE")},
				Meta:       MetaConfiguration{Image: os.Getenv("META_IMAGE")},
				Database: DatabaseConfiguration{
					Exporter: ExporterConfiguration{Image: os.Getenv("PSQL_EXPORTER_IMAGE")},
					Resources: ResourcesWithPersistentVolume{
						VolumeAccessMode:   os.Getenv("DATABASE_VOLUME_ACCESS_MODE"),
						VolumeStorageClass: os.Getenv("DATABASE_STORAGE_CLASS"),
						VolumeName:         os.Getenv("DATABASE_VOLUME_NAME"),
					},
				},
				Server: ServerConfiguration{
					Image: os.Getenv("SERVER_IMAGE"),
				},
			},
		},
	}

	if err := mergo.Merge(config, imgEnv, mergo.WithOverride); err != nil {
		return err
	}

	config.DevSupport = setBoolFromEnv("DEV_SUPPORT", config.DevSupport)
	config.Syndesis.Components.Server.Features.TestSupport = setBoolFromEnv("TEST_SUPPORT", config.Syndesis.Components.Server.Features.TestSupport)

	return nil
}

// Return the value of a config given its default value and an environment
// variable.
func setBoolFromEnv(env string, current bool) bool {
	var result bool
	if varFromEnv := os.Getenv(env); varFromEnv != "" {
		result = varFromEnv == "true"
	} else {
		result = current
	}

	return result
}

// Return the value of a config given its default value and an environment
// variable.
func setIntFromEnv(env string, current int) int {
	if varFromEnv := os.Getenv(env); varFromEnv != "" {
		if result, err := strconv.Atoi(varFromEnv); err == nil {
			return result
		}
	}

	return current
}

// Replace default values with those from custom resource
func (config *Config) setSyndesisFromCustomResource(syndesis *v1beta1.Syndesis) error {
	c := SyndesisConfig{}
	jsonProperties, err := json.Marshal(syndesis.Spec)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonProperties, &c); err != nil {
		return err
	}

	if err := mergo.Merge(&config.Syndesis, c, mergo.WithOverride); err != nil {
		return err
	}
	return nil
}

// Generate random expressions for passwords and secrets
func (config *Config) generatePasswords() {

	if config.OpenShiftOauthClientSecret == "" {
		config.OpenShiftOauthClientSecret = generatePassword(64)
	}

	if config.Syndesis.Components.Database.Password == "" {
		config.Syndesis.Components.Database.Password = generatePassword(16)
	}

	if config.Syndesis.Components.Database.SampledbPassword == "" {
		config.Syndesis.Components.Database.SampledbPassword = generatePassword(16)
	}

	if config.Syndesis.Components.Oauth.CookieSecret == "" {
		config.Syndesis.Components.Oauth.CookieSecret = generatePassword(32)
	}

	if config.Syndesis.Components.Server.SyndesisEncryptKey == "" {
		config.Syndesis.Components.Server.SyndesisEncryptKey = generatePassword(64)
	}

	if config.Syndesis.Components.Server.ClientStateAuthenticationKey == "" {
		config.Syndesis.Components.Server.ClientStateAuthenticationKey = generatePassword(32)
	}

	if config.Syndesis.Components.Server.ClientStateEncryptionKey == "" {
		config.Syndesis.Components.Server.ClientStateEncryptionKey = generatePassword(32)
	}
}

func generatePassword(size int) string {
	alphabet := make([]rune, (26*2)+10)
	i := 0
	for c := 'a'; c <= 'z'; c++ {
		alphabet[i] = c
		i += 1
	}
	for c := 'A'; c <= 'Z'; c++ {
		alphabet[i] = c
		i += 1
	}
	for c := '0'; c <= '9'; c++ {
		alphabet[i] = c
		i += 1
	}

	result := make([]rune, size)
	for i := 0; i < size; i++ {
		result[i] = alphabet[random.Intn(len(alphabet))]
	}
	s := string(result)
	return s
}

// Needed for the first run after upgrade, due to compatibilities with old
// secret format
// TODO: Delete for 1.10
func parseConfigurationBlob(blob []byte) map[string]string {
	strs := strings.Split(string(blob), "\n")
	configs := make(map[string]string, 0)
	for _, conf := range strs {
		conf := strings.Trim(conf, " \r\t")
		if conf == "" {
			continue
		}
		kv := strings.SplitAfterN(conf, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimRight(kv[0], "=")
			value := kv[1]
			configs[key] = value
		}
	}
	return configs
}

// TODO: Delete for 1.10
func getSyndesisEnvVarsFromOpenShiftNamespace(secret *corev1.Secret) (map[string]string, error) {
	if envBlob, present := secret.Data["params"]; present {
		return parseConfigurationBlob(envBlob), nil
	} else {
		return nil, errors.New("no configuration found")
	}
}
