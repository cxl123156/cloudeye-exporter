package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

const MinimumResourceInfoSyncInterval = 10

var tagRegexp *regexp.Regexp

func init() {
	var err error
	tagRegexp, err = regexp.Compile("^([a-z]|[A-Z]){1}([a-z]|[A-Z]|_)*$")
	if err != nil {
		logs.Logger.Error("init tag regexp error: %s", err.Error())
	}
}

type serversInfo struct {
	TTL           int64
	LabelInfo     map[string]labelInfo
	FilterMetrics []model.MetricInfoList
	sync.Mutex
}

type labelInfo struct {
	Name  []string
	Value []string
}

type RmsInfo struct {
	Id   string
	Name string
	EpId string
	Tags map[string]string
}

func GetResourceKeyFromMetricInfo(metric model.MetricInfoList) string {
	sort.Slice(metric.Dimensions, func(i, j int) bool {
		return metric.Dimensions[i].Name < metric.Dimensions[j].Name
	})
	dimValuesList := make([]string, 0, len(metric.Dimensions))
	for _, dim := range metric.Dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func GetResourceKeyFromMetricData(metric model.BatchMetricData) string {
	// DMS实例其他维度不需要适配资源标签，只匹配实例信息
	if *metric.Namespace == "SYS.DMS" {
		return getDmsResourceKey(metric)
	}
	if *metric.Namespace == "AGT.ECS" || *metric.Namespace == "SERVICE.BMS" {
		return getServerResourceKey(metric)
	}
	if *metric.Namespace == "SYS.MRS" {
		return getMrsResourceKey(metric)
	}
	sort.Slice(*metric.Dimensions, func(i, j int) bool {
		return (*metric.Dimensions)[i].Name < (*metric.Dimensions)[j].Name
	})
	dimValuesList := make([]string, 0, len(*metric.Dimensions))
	for _, dim := range *metric.Dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func getServerResourceKey(metric model.BatchMetricData) string {
	for _, dim := range *metric.Dimensions {
		if dim.Name == "instance_id" {
			return dim.Value
		}
	}
	return ""
}

func getDmsResourceKey(metric model.BatchMetricData) string {
	for _, dim := range *metric.Dimensions {
		if dim.Name == "kafka_instance_id" || dim.Name == "rabbitmq_instance_id" || dim.Name == "reliablemq_instance_id" {
			return dim.Value
		}
	}
	return ""
}

func getMrsResourceKey(metric model.BatchMetricData) string {
	for _, dim := range *metric.Dimensions {
		if dim.Name == "cluster_id" {
			return dim.Value
		}
	}
	return ""
}

func getEndpoint(server, version string) string {
	if endpoint, ok := endpointConfig[server]; ok {
		return endpoint
	}
	return fmt.Sprintf("https://%s/%s", strings.Replace(host, "iam", server, 1), version)
}

// 标签只允许大写字母，小写字母和下划线，过滤tags中有效的tag
func getTags(tags map[string]string) ([]string, []string) {
	var keys, values []string
	for key, value := range tags {
		valid := tagRegexp.MatchString(key)
		if !valid {
			continue
		}
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func fmtTags(tagInfo interface{}) map[string]string {
	bytes, err := json.Marshal(tagInfo)
	if err != nil {
		return nil
	}
	var tags []Tag
	err = json.Unmarshal(bytes, &tags)
	if err != nil {
		return nil
	}
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}

type ResourceBaseInfo struct {
	ID   string
	Name string
	EpId string
	Tags map[string]string
}

func getDimsNameKey(dims []model.MetricsDimension) string {
	dimsNamesList := make([]string, 0, len(dims))
	for _, dim := range dims {
		dimsNamesList = append(dimsNamesList, dim.Name)
	}
	return strings.Join(dimsNamesList, ",")
}

func getDimsValueKey(dims []model.MetricsDimension) string {
	dimsValuesList := make([]string, 0, len(dims))
	for _, dim := range dims {
		dimsValuesList = append(dimsValuesList, dim.Value)
	}
	return strings.Join(dimsValuesList, ",")
}

func buildSingleDimensionMetrics(metricNames []string, namespace, dimName, dimValue string) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: []model.MetricsDimension{
				{
					Name:  dimName,
					Value: dimValue,
				},
			},
		}
	}
	return filterMetrics
}

func buildDimensionMetrics(metricNames []string, namespace string, dimensions []model.MetricsDimension) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: dimensions,
		}
	}
	return filterMetrics
}

func getHcClient(endpoint string) *core.HcHttpClient {
	return core.NewHcHttpClient(impl.NewDefaultHttpClient(config.DefaultHttpConfig().WithIgnoreSSLVerification(true))).
		WithCredential(basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithEndpoints([]string{endpoint})
}

func genDefaultReqDefWithOffsetAndLimit(path string, response interface{}) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath(path).
		WithResponse(response).WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Offset").WithJsonTag("offset").WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Limit").WithJsonTag("limit").WithLocationType(def.Query))
	return reqDefBuilder.Build()
}

func getDefaultString(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

func fmtResourceProperties(properties map[string]interface{}, value interface{}) error {
	bytes, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

func getResourcesBaseInfoFromRMS(provider, resourceType string) ([]ResourceBaseInfo, error) {
	resp, err := listResources(provider, resourceType)
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of %s.%s, error: %s", provider, resourceType, err.Error())
		return nil, err
	}
	services := make([]ResourceBaseInfo, len(resp))
	for index, resource := range resp {
		services[index].ID = *resource.Id
		services[index].Name = *resource.Name
		services[index].EpId = *resource.EpId
		services[index].Tags = resource.Tags
	}
	return services, nil
}

func GetResourceInfoExpirationTime() time.Duration {
	intervalMinutes := CloudConf.Global.ResourceSyncIntervalMinutes
	if intervalMinutes <= MinimumResourceInfoSyncInterval {
		return MinimumResourceInfoSyncInterval * time.Minute
	}
	return time.Duration(intervalMinutes) * time.Minute
}
