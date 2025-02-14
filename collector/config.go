package collector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	v3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"gopkg.in/yaml.v3"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type CloudAuth struct {
	ProjectName string `yaml:"project_name"`
	ProjectID   string `yaml:"project_id"`
	DomainName  string `yaml:"domain_name"`
	// 建议您优先使用ReadMe文档中 4.1章节的方式，使用脚本将AccessKey,SecretKey解密后传入，避免在配置文件中明文配置AK SK导致信息泄露
	AccessKey string `yaml:"access_key"`
	Region    string `yaml:"region"`
	SecretKey string `yaml:"secret_key"`
	AuthURL   string `yaml:"auth_url"`
}

type Global struct {
	Port                        string `yaml:"port"`
	Prefix                      string `yaml:"prefix"`
	MetricPath                  string `yaml:"metric_path"`
	EpsInfoPath                 string `yaml:"eps_path"`
	MaxRoutines                 int    `yaml:"max_routines"`
	ScrapeBatchSize             int    `yaml:"scrape_batch_size"`
	ResourceSyncIntervalMinutes int    `yaml:"resource_sync_interval_minutes"`
	EpIds                       string `yaml:"ep_ids"`
	MetricsConfPath             string `yaml:"metrics_conf_path"`
	LogsConfPath                string `yaml:"logs_conf_path"`
	EndpointsConfPath           string `yaml:"endpoints_conf_path"`
	IgnoreSSLVerify             bool   `yaml:"ignore_ssl_verify"`

	// 用户配置的proxy信息
	HttpSchema string `yaml:"proxy_schema"`
	HttpHost   string `yaml:"proxy_host"`
	HttpPort   int    `yaml:"proxy_port"`
	UserName   string `yaml:"proxy_username"`
	Password   string `yaml:"proxy_password"`

	// CN列表，用于校验https证书链中的DNS名称
	ClientCN string `yaml:"client_cn"`
}

type CloudConfig struct {
	Auth   CloudAuth `yaml:"auth"`
	Global Global    `yaml:"global"`
}

var CloudConf CloudConfig
var SecurityMod bool
var HttpsEnabled bool
var ProxyEnabled bool
var TmpAK string
var TmpSK string
var TmpProxyUserName string
var TmpProxyPassword string

func InitCloudConf(file string) error {
	realPath, err := NormalizePath(file)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(realPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &CloudConf)
	if err != nil {
		return err
	}

	SetDefaultConfigValues(&CloudConf)

	err = InitConfig()
	if err != nil {
		return err
	}
	return err
}

func NormalizePath(path string) (string, error) {
	relPath, err := filepath.Abs(path) // 对文件路径进行标准化
	if err != nil {
		return "", err
	}
	relPath = strings.Replace(relPath, "\\", "/", -1)
	match, err := regexp.MatchString("[!;<>&|$\n`\\\\]", relPath)
	if match || err != nil {
		return "", errors.New("match path error")
	}
	return relPath, nil
}

func SetDefaultConfigValues(config *CloudConfig) {
	if config.Global.Port == "" {
		config.Global.Port = ":8087"
	}

	if config.Global.MetricPath == "" {
		config.Global.MetricPath = "/metrics"
	}

	if config.Global.EpsInfoPath == "" {
		config.Global.EpsInfoPath = "/eps-info"
	}

	if config.Global.Prefix == "" {
		config.Global.Prefix = "huaweicloud"
	}

	if config.Global.MaxRoutines == 0 {
		config.Global.MaxRoutines = 5
	}

	if config.Global.ScrapeBatchSize == 0 {
		config.Global.ScrapeBatchSize = 300
	}

	if config.Global.ResourceSyncIntervalMinutes <= 0 {
		config.Global.ResourceSyncIntervalMinutes = 180
	}

	if config.Global.MetricsConfPath == "" {
		config.Global.MetricsConfPath = "./metric.yml"
	}

	if config.Global.LogsConfPath == "" {
		config.Global.LogsConfPath = "./logs.yml"
	}

	if config.Global.EndpointsConfPath == "" {
		config.Global.EndpointsConfPath = "./endpoints.yml"
	}
}

type MetricConf struct {
	Resource      string              `yaml:"resource"`
	DimMetricName map[string][]string `yaml:"dim_metric_name"`
}

var metricConf map[string]MetricConf

func InitMetricConf() error {
	metricConf = make(map[string]MetricConf)
	realPath, err := NormalizePath(CloudConf.Global.MetricsConfPath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(realPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &metricConf)
}

func getMetricConfigMap(namespace string) map[string][]string {
	if conf, ok := metricConf[namespace]; ok {
		return conf.DimMetricName
	}
	return nil
}

func getResourceFromRMS(namespace string) bool {
	if conf, ok := metricConf[namespace]; ok {
		return conf.Resource == "RMS" || conf.Resource == "rms"
	}
	return false
}

type Config struct {
	AccessKey        string
	SecretKey        string
	DomainID         string
	DomainName       string
	EndpointType     string
	IdentityEndpoint string
	Region           string
	ProjectID        string
	ProjectName      string
	UserID           string
}

var conf = &Config{}

func InitConfig() error {
	conf.IdentityEndpoint = CloudConf.Auth.AuthURL
	conf.ProjectName = CloudConf.Auth.ProjectName
	conf.ProjectID = CloudConf.Auth.ProjectID
	conf.DomainName = CloudConf.Auth.DomainName
	conf.Region = CloudConf.Auth.Region
	// 安全模式下，ak/sk通过用户交互获取，避免明文方式存在于存储介质中
	if SecurityMod {
		conf.AccessKey = TmpAK
		conf.SecretKey = TmpSK
	} else {
		conf.AccessKey = CloudConf.Auth.AccessKey
		conf.SecretKey = CloudConf.Auth.SecretKey
	}

	if conf.ProjectID == "" && conf.ProjectName == "" {
		fmt.Printf("Init config error: ProjectID or ProjectName must setting.")
		return errors.New("init config error: ProjectID or ProjectName must setting")
	}
	req, err := http.NewRequest("GET", conf.IdentityEndpoint, nil)
	if err != nil {
		fmt.Printf("Auth url is invalid.")
		return err
	}
	host = req.Host

	if conf.ProjectID == "" {
		resp, err := getProjectInfo()
		if err != nil {
			fmt.Printf("Get project info error: %s", err.Error())
			return err
		}
		if len(*resp.Projects) == 0 {
			fmt.Printf("Project info is empty")
			return errors.New("project info is empty")
		}

		projects := *resp.Projects
		conf.ProjectID = projects[0].Id
		conf.DomainID = projects[0].DomainId
	}
	return nil
}

func getProjectInfo() (*model.KeystoneListProjectsResponse, error) {
	iamclient := v3.NewIamClient(
		v3.IamClientBuilder().
			WithEndpoint(conf.IdentityEndpoint).
			WithCredential(
				global.NewCredentialsBuilder().
					WithAk(conf.AccessKey).
					WithSk(conf.SecretKey).
					Build()).
			WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
			Build())
	return iamclient.KeystoneListProjects(&model.KeystoneListProjectsRequest{Name: &conf.ProjectName})
}

var endpointConfig map[string]string

func InitEndpointConfig(path string) {
	realPath, err := NormalizePath(path)
	if err != nil {
		logs.Logger.Errorf("Normalize endpoint config err: %s", err.Error())
		return
	}

	context, err := ioutil.ReadFile(realPath)
	if err != nil {
		logs.Logger.Infof("Invalid endpoint config path, default config will be used instead")
		return
	}
	err = yaml.Unmarshal(context, &endpointConfig)
	if err != nil {
		logs.Logger.Errorf("Init endpoint config error: %s", err.Error())
	}
}
