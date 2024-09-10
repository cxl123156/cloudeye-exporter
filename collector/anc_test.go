package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestANCInfo_GetResourceInfo(t *testing.T) {
	metricConf = map[string]MetricConf{
		"SYS.ANC": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"anc_anc_id":                           {"network_incoming_bits_rate"},
				"anc_anc_id,anc_clientassociation_id":  {"vpc_connections"},
				"anc_anc_id,anc_serviceassociation_id": {"service_inbound_traffic"},
				"anc_service_id":                       {"service_inbound_traffic"},
				"anc_service_id,anc_serviceregion_id":  {"service_inbound_traffic"},
				"anc_membergroup_id":                   {"total_server_count"},
			},
		},
	}
	patches := gomonkey.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			MetricName: "network_incoming_bits_rate",
			Dimensions: []model.MetricsDimension{
				{Name: "anc_anc_id", Value: "0cbcebbf-7732-4d50-8186-45cdf2fec1f3"},
			},
		},
		{
			MetricName: "service_inbound_traffic",
			Dimensions: []model.MetricsDimension{
				{Name: "anc_service_id", Value: "440a1bf6-68cc-4a6a-b286-e89811a99b61"},
			},
		},
		{
			MetricName: "total_server_count",
			Dimensions: []model.MetricsDimension{
				{Name: "anc_membergroup_id", Value: "139c8a1c-3b5d-4a97-8095-1b8aeed2b93e"},
			},
		},
		{
			MetricName: "vpc_connections",
			Dimensions: []model.MetricsDimension{
				{Name: "anc_anc_id", Value: "0cbcebbf-7732-4d50-8186-45cdf2fec1f3"},
				{Name: "anc_clientassociation_id", Value: "2da0f177-3294-4b52-ae8b-719e899ab5a5"},
			},
		},
		{
			MetricName: "service_inbound_traffic",
			Dimensions: []model.MetricsDimension{
				{Name: "anc_anc_id", Value: "0cbcebbf-7732-4d50-8186-45cdf2fec1f3"},
				{Name: "anc_serviceassociation_id", Value: "04c73e29-848a-4a45-8dd2-6dc1b79c2832"},
			},
		},
		{
			MetricName: "service_inbound_traffic",
			Dimensions: []model.MetricsDimension{
				{Name: "anc_service_id", Value: "440a1bf6-68cc-4a6a-b286-e89811a99b61"},
				{Name: "anc_serviceregion_id", Value: "440a1bf6-68cc-4a6a-b286-e89811a99b61_cn-southwest-242"},
			},
		},
	}, nil)
	defer patches.Reset()
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "0cbcebbf-7732-4d50-8186-45cdf2fec1f3",
					Name: "demo-anc",
					EpId: "0",
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "440a1bf6-68cc-4a6a-b286-e89811a99b61",
					Name: "demo-anc-service",
					EpId: "0",
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "139c8a1c-3b5d-4a97-8095-1b8aeed2b93e",
					Name: "demo-anc-member-group",
					EpId: "0",
				},
			}, nil},
		},
	}
	patches.ApplyFuncSeq(getResourcesBaseInfoFromRMS, outputs)
	var ancInfo ANCInfo
	labels, metrics := ancInfo.GetResourceInfo()
	assert.Equal(t, 6, len(labels))
	assert.Equal(t, 6, len(metrics))
}
