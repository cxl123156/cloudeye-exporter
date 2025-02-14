package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	cdm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdm/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdm/v1/model"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var cdmInfo serversInfo

type CDMInfo struct{}

func (getter CDMInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	cdmInfo.Lock()
	defer cdmInfo.Unlock()
	if cdmInfo.LabelInfo == nil || time.Now().Unix() > cdmInfo.TTL {
		if metricNames, ok := getMetricConfigMap("SYS.CDM")["instance_id"]; ok {
			clusters := listClusters()
			if clusters == nil {
				return cdmInfo.LabelInfo, cdmInfo.FilterMetrics
			}
			for _, cluster := range *clusters {
				for _, instance := range *cluster.Instances {
					metrics := buildSingleDimensionMetrics(metricNames, "SYS.CDM", "instance_id", instance.Id)
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"clusterId", "clusterName", "instanceName", "trafficIp", "manage_ip", "private_ip"},
						Value: []string{cluster.Id, cluster.Name, instance.Name, getDefaultString(instance.TrafficIp), getDefaultString(instance.ManageIp), getDefaultString(instance.PrivateIp)},
					}
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}
		cdmInfo.LabelInfo = resourceInfos
		cdmInfo.FilterMetrics = filterMetrics
		cdmInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return cdmInfo.LabelInfo, cdmInfo.FilterMetrics
}

func getCDMClient() *cdm.CdmClient {
	return cdm.NewCdmClient(cdm.CdmClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("cdm", "v1.1")).Build())
}

func listClusters() *[]model.Clusters {
	response, err := getCDMClient().ListClusters(&model.ListClustersRequest{})
	if err != nil {
		logs.Logger.Errorf("list cdm clusters error: %s", err.Error())
		return nil
	}
	return response.Clusters
}
