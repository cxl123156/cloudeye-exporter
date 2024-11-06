package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	clouttable "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cloudtable/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cloudtable/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var cloudTableInfo serversInfo

type CloudTableInfo struct{}

func getCloudTableClient() *clouttable.CloudTableClient {
	return clouttable.NewCloudTableClient(clouttable.CloudTableClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("cloudtable", "v2")).Build())
}

func (ct CloudTableInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	cloudTableInfo.Lock()
	defer cloudTableInfo.Unlock()
	if cloudTableInfo.LabelInfo == nil || time.Now().Unix() > cloudTableInfo.TTL {
		clusters := GetClusterInfo()
		for _, cluster := range clusters {
			info := labelInfo{
				Name:  []string{"clusterName"},
				Value: []string{*cluster.ClusterName},
			}
			resourceInfos[*cluster.ClusterId] = info
		}
		allMetrics, err := listAllMetrics("SYS.CloudTable")
		if err != nil {
			logs.Logger.Errorf("[%s] Get all metrics of SYS.CloudTable error: %s", err.Error())
			return cloudTableInfo.LabelInfo, cloudTableInfo.FilterMetrics
		}
		for _, metric := range allMetrics {
			if info, ok := resourceInfos[metric.Dimensions[0].Value]; ok {
				resourceInfos[GetResourceKeyFromMetricInfo(metric)] = info
			}
		}
		cloudTableInfo.LabelInfo = resourceInfos
		cloudTableInfo.FilterMetrics = allMetrics
		cloudTableInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return cloudTableInfo.LabelInfo, cloudTableInfo.FilterMetrics
}

func GetClusterInfo() []model.ClusterDetail {
	cloudTableClusterLimit := int32(100)
	cloudTableClusterOffset := int32(0)
	request := &model.ListClustersRequest{Limit: &cloudTableClusterLimit, Offset: &cloudTableClusterOffset}
	var clusters []model.ClusterDetail
	for {
		response, err := getCloudTableClient().ListClusters(request)
		if err != nil {
			logs.Logger.Errorf("list cloud table clusters error: %s, limit: %d, offset: %d", err.Error(),
				*request.Limit, *request.Offset)
			return clusters
		}
		tempClusters := *response.Clusters
		if len(tempClusters) == 0 {
			break
		}
		clusters = append(clusters, tempClusters...)
		*request.Offset += cloudTableClusterLimit
	}
	return clusters
}
