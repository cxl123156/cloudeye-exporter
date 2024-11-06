package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	dnsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var dnsServerInfo serversInfo

type DNSInfo struct{}

func (getter DNSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	filterMetrics := make([]model.MetricInfoList, 0)
	resourceInfos := map[string]labelInfo{}
	dnsServerInfo.Lock()
	defer dnsServerInfo.Unlock()

	if dnsServerInfo.LabelInfo == nil || time.Now().Unix() > dnsServerInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DNS")

		// get record set
		metricNames, ok := sysConfigMap["dns_recordset_id"]
		if !ok {
			logs.Logger.Warnf("Metric config is empty of SYS.DNS, dim_metric_name is dns_recordset_id")
			return resourceInfos, filterMetrics
		}

		var allRecordSets []dnsModel.ListRecordSetsWithTags
		publicRecordSets, err := getAllRecordSets("public")
		if err != nil {
			logs.Logger.Errorf("Failed to get public record sets, error is: %s", err.Error())
			return resourceInfos, filterMetrics
		}
		allRecordSets = append(allRecordSets, publicRecordSets...)

		privateRecordSets, err := getAllRecordSets("private")
		if err != nil {
			logs.Logger.Errorf("Failed to get private record sets, error is: %s", err.Error())
			return resourceInfos, filterMetrics
		}
		allRecordSets = append(allRecordSets, privateRecordSets...)

		for _, record := range allRecordSets {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DNS", "dns_recordset_id", *record.Id)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "type"},
				Value: []string{*record.Name, *record.Type},
			}
			tagMap := fmtTags(*record.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}

		// get public zone
		metricNames, ok = sysConfigMap["dns_zone_id"]
		if !ok {
			logs.Logger.Warnf("Metric config is empty of SYS.DNS, dim_metric_name is dns_zone_id")
			return resourceInfos, filterMetrics
		}
		publicZones, err := getAllPublicZones()
		if err != nil {
			logs.Logger.Errorf("Failed to get all public zones, error is: %s", err.Error())
			return resourceInfos, filterMetrics
		}
		for _, zone := range publicZones {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DNS", "dns_zone_id", *zone.Id)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "zone_type", "epId"},
				Value: []string{*zone.Name, *zone.ZoneType, *zone.EnterpriseProjectId},
			}
			tagMap := fmtTags(*zone.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
		dnsServerInfo.LabelInfo = resourceInfos
		dnsServerInfo.FilterMetrics = filterMetrics
		dnsServerInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return dnsServerInfo.LabelInfo, dnsServerInfo.FilterMetrics
}

func getAllRecordSets(zoneType string) ([]dnsModel.ListRecordSetsWithTags, error) {
	limit := int32(500)
	offset := int32(0)
	var recordSets []dnsModel.ListRecordSetsWithTags

	for {
		request := &dnsModel.ListRecordSetsRequest{ZoneType: &zoneType, Limit: &limit, Offset: &offset}
		response, err := getDnsClient().ListRecordSets(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all record sets, type is: %s, error is: %s", zoneType, err.Error())
			return nil, err
		}
		if len(*response.Recordsets) == 0 {
			break
		}
		recordSets = append(recordSets, *response.Recordsets...)
		*request.Offset += limit
	}
	return recordSets, nil
}

func getAllPublicZones() ([]dnsModel.PublicZoneResp, error) {
	limit := int32(500)
	offset := int32(0)

	var publicZones []dnsModel.PublicZoneResp
	public := "public"
	for {
		request := &dnsModel.ListPublicZonesRequest{Type: &public, Limit: &limit, Offset: &offset}
		response, err := getDnsClient().ListPublicZones(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all public zones, error is: %s", err.Error())
			return nil, err
		}
		if len(*response.Zones) == 0 {
			break
		}
		publicZones = append(publicZones, *response.Zones...)
		*request.Offset += limit
	}
	return publicZones, nil
}

func getDnsClient() *dns.DnsClient {
	return dns.NewDnsClient(dns.DnsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("dns", "v2")).Build())
}
