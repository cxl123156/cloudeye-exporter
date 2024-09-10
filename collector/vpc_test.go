package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestVpcGetResourceInfo(t *testing.T) {
	metricConf = make(map[string]MetricConf)
	metricConf["SYS.VPC"] = MetricConf{
		Resource: "rms",
		DimMetricName: map[string][]string{
			"publicip_id": {"up_bandwidth"}, "bandwidth_id": {"up_bandwidth"},
		},
	}
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(listResources, func(provider, resourceType string) ([]model.ResourceEntity, error) {
		if resourceType == "bandwidths" {
			bandwidthName := "bandwidth-743a"
			bandwidthID := "3926d9fe-da8f-4005-9027-f0871dae9f0b"
			bandwidthEpID := "0"
			return []model.ResourceEntity{
				{
					Name: &bandwidthName,
					Id:   &bandwidthID,
					EpId: &bandwidthEpID,
					Tags: map[string]string{},
					Properties: map[string]interface{}{
						"chargeMode": "traffic",
					},
				},
			}, nil
		} else {
			publicIpName := "100.xx.xxx.xxx"
			publicIpID := "3926d9fe-da8f-4005-9027-f0871dae9f0a"
			publicIpEpID := "0"
			return []model.ResourceEntity{
				{
					Name: &publicIpName,
					Id:   &publicIpID,
					EpId: &publicIpEpID,
					Tags: map[string]string{},
					Properties: map[string]interface{}{
						"bandwidth": map[string]string{
							"name":       "bandwidth-743a",
							"chargeMode": "traffic",
						},
					},
				},
			}, nil
		}
	})
	defer patches.Reset()

	var vpcGetter VPCInfo
	vpcInfo.LabelInfo = nil
	labels, metrics := vpcGetter.GetResourceInfo()
	metricConf = make(map[string]MetricConf)
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, 2, len(metrics))
}
