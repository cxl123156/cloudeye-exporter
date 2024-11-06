package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestEGInfo_GetResourceInfo(t *testing.T) {
	metricConf = map[string]MetricConf{
		"SYS.EG": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"source_name":               {"event_num"},
				"channel_id":                {"pub_process_time"},
				"streaming_id":              {"streaming_live_status"},
				"subscription_id":           {"sub_num"},
				"subscription_id,target_id": {"sub_failed_rate"},
			},
		},
	}
	patches := gomonkey.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			MetricName: "event_num",
			Dimensions: []model.MetricsDimension{
				{Name: "source_name", Value: "test-source-name"},
			},
		},
		{
			MetricName: "pub_process_time",
			Dimensions: []model.MetricsDimension{
				{Name: "channel_id", Value: "440a1bf6-68cc-4a6a-b286-e89811a99b61"},
			},
		},
		{
			MetricName: "streaming_live_status",
			Dimensions: []model.MetricsDimension{
				{Name: "streaming_id", Value: "139c8a1c-3b5d-4a97-8095-1b8aeed2b93e"},
			},
		},
		{
			MetricName: "sub_num",
			Dimensions: []model.MetricsDimension{
				{Name: "subscription_id", Value: "0cbcebbf-7732-4d50-8186-45cdf2fec1f3"},
			},
		},
		{
			MetricName: "sub_failed_rate",
			Dimensions: []model.MetricsDimension{
				{Name: "subscription_id", Value: "0cbcebbf-7732-4d50-8186-45cdf2fec1f3"},
				{Name: "target_id", Value: "04c73e29-848a-4a45-8dd2-6dc1b79c2832"},
			},
		},
	}, nil)
	defer patches.Reset()
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "test-source-name",
					Name: "test-source-name",
					EpId: "0",
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "440a1bf6-68cc-4a6a-b286-e89811a99b61",
					Name: "test-channel-name",
					EpId: "0",
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "139c8a1c-3b5d-4a97-8095-1b8aeed2b93e",
					Name: "test-stream-name",
					EpId: "0",
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{[]ResourceBaseInfo{
				{
					ID:   "0cbcebbf-7732-4d50-8186-45cdf2fec1f3",
					Name: "test-subscription-name",
					EpId: "0",
				},
			}, nil},
		},
	}
	patches.ApplyFuncSeq(getResourcesBaseInfoFromRMS, outputs)
	var egInfo EgInfo
	labels, metrics := egInfo.GetResourceInfo()
	assert.Equal(t, 5, len(labels))
	assert.Equal(t, 5, len(metrics))
}
