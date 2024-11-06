package collector

import (
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	egInfo         serversInfo
	egDimConfigMap = map[string]map[string]string{
		"channel_id": {
			"resourceType":      "channels",
			"providerId":        "eg",
			"nameInSubResource": "channel_name",
		},
		"source_name": {
			"resourceType":      "sources",
			"providerId":        "eg",
			"nameInSubResource": "source_name",
		},
		"subscription_id": {
			"resourceType":      "subscriptions",
			"providerId":        "eg",
			"nameInSubResource": "subscription_name",
		},
		"streaming_id": {
			"resourceType":      "streamings",
			"providerId":        "eg",
			"nameInSubResource": "streaming_name",
		},
	}
)

type EgInfo struct{}

func (getter EgInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	egInfo.Lock()
	defer egInfo.Unlock()
	if egInfo.LabelInfo == nil || time.Now().Unix() > egInfo.TTL {
		resourceInfos := map[string]labelInfo{}
		allMetrics, err := listAllMetrics("SYS.EG")
		if err != nil {
			logs.Logger.Errorf("Get all eg metrics: %s", err.Error())
			return egInfo.LabelInfo, egInfo.FilterMetrics
		}
		egResourceMetricMap := getEgSubResourceMetrics(allMetrics)
		getEgResourceInfoFromRms("source_name", resourceInfos, egResourceMetricMap)
		getEgResourceInfoFromRms("channel_id", resourceInfos, egResourceMetricMap)
		getEgResourceInfoFromRms("subscription_id", resourceInfos, egResourceMetricMap)
		getEgResourceInfoFromRms("streaming_id", resourceInfos, egResourceMetricMap)
		egInfo.LabelInfo = resourceInfos
		var filteredMetrics []cesmodel.MetricInfoList
		for _, metrics := range allMetrics {
			if _, ok := resourceInfos[GetResourceKeyFromMetricInfo(metrics)]; ok {
				filteredMetrics = append(filteredMetrics, metrics)
			}
		}
		egInfo.FilterMetrics = filteredMetrics
		egInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return egInfo.LabelInfo, egInfo.FilterMetrics
}

func getEgResourceInfoFromRms(dimName string, resourceInfos map[string]labelInfo, allMetricMap map[string][]cesmodel.MetricInfoList) {
	dimConfig, ok := egDimConfigMap[dimName]
	if !ok {
		return
	}
	rmsResourceInfos, err := getResourcesBaseInfoFromRMS(dimConfig["providerId"], dimConfig["resourceType"])
	if err != nil {
		return
	}
	for _, rmsResourceInfo := range rmsResourceInfos {
		info := labelInfo{
			Name:  []string{"name", "epId"},
			Value: []string{rmsResourceInfo.Name, rmsResourceInfo.EpId},
		}
		keys, values := getTags(rmsResourceInfo.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[rmsResourceInfo.ID] = info

		// 设置子维度指标资源关联
		subMetricList, ok := allMetricMap[rmsResourceInfo.ID]
		if !ok {
			continue
		}
		for _, subMetric := range subMetricList {
			resourceInfos[GetResourceKeyFromMetricInfo(subMetric)] = labelInfo{
				Name:  []string{dimConfig["nameInSubResource"], "epId"},
				Value: []string{rmsResourceInfo.Name, rmsResourceInfo.EpId},
			}
		}
	}
}

func getEgSubResourceMetrics(allMetrics []cesmodel.MetricInfoList) map[string][]cesmodel.MetricInfoList {
	resultMap := make(map[string][]cesmodel.MetricInfoList)
	for _, metric := range allMetrics {
		if len(metric.Dimensions) < 2 {
			continue
		}
		firstDimValue := metric.Dimensions[0].Value
		subMetricList, ok := resultMap[firstDimValue]
		if ok {
			subMetricList = append(subMetricList, metric)
		} else {
			subMetricList = []cesmodel.MetricInfoList{metric}
		}
		resultMap[firstDimValue] = subMetricList
	}
	return resultMap
}
