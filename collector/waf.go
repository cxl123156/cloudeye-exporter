package collector

import (
	"net/http"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	waf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
	wafModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var wafInfo serversInfo

type WAFInfo struct{}

func (getter WAFInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	wafInfo.Lock()
	defer wafInfo.Unlock()
	if wafInfo.LabelInfo == nil || time.Now().Unix() > wafInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.WAF")
		wafInstances, err := getAllWafInstancesFromRMS()
		if err != nil {
			logs.Logger.Errorf("Failed to get all waf instances, error: %s", err.Error())
			return nil, nil
		}

		for _, instance := range wafInstances {
			if metricNames, ok := sysConfigMap["waf_instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.WAF", "waf_instance_id", instance.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId"},
					Value: []string{instance.Name, instance.EpId},
				}
				keys, values := getTags(instance.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		premiumWafInstances := getAllPremiumWafInstances()
		if premiumWafInstances == nil {
			return nil, nil
		}
		for _, instance := range premiumWafInstances {
			if metricNames, ok := sysConfigMap["instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.WAF", "instance_id", *instance.Id)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name"},
					Value: []string{*instance.InstanceName},
				}
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		wafInfo.LabelInfo = resourceInfos
		wafInfo.FilterMetrics = filterMetrics
		wafInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return wafInfo.LabelInfo, wafInfo.FilterMetrics
}

func getAllWafInstancesFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("waf", "instance")
}

func getWAFClient() *waf.WafClient {
	return waf.NewWafClient(waf.WafClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("waf", "v1")).Build())
}

func getAllPremiumWafInstances() []wafModel.ListInstance {
	resp, err := getWAFClient().ListInstance(&wafModel.ListInstanceRequest{})
	if resp.HttpStatusCode != http.StatusOK {
		logs.Logger.Errorf("Get all premiumWafInstances HttpStatusCode is %d", resp.HttpStatusCode)
		return nil
	}
	if err != nil {
		logs.Logger.Errorf("Get all premiumWafInstances err, err is : %s", err.Error())
		return nil
	}
	return *resp.Items
}

func (getter WAFInfo) resetResourceInfo() {
	wafInfo.LabelInfo = nil
}
