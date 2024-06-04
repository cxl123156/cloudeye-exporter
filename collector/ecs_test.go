package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestECSInfo_GetResourceInfo(t *testing.T) {
	patches :=
		gomonkey.ApplyFuncReturn(getAllServer, []EcsInstancesInfo{
			{
				ResourceBaseInfo: ResourceBaseInfo{
					ID:   "c53dbd97-c4ae-4ef7-8d12-fae2d3d38a80",
					Name: "nodelete-DRS-autotest-datamode1-test2",
					Tags: map[string]string{"wukong": "000"},
					EpId: "0",
				},
				IP: "192.168.20.53,100.93.4.203",
			},
		}, nil)
	patches.ApplyFuncReturn(getAllServerFromRMS, []EcsInstancesInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "c53dbd97-c4ae-4ef7-8d12-fae2d3d38a80",
				Name: "nodelete-DRS-autotest-datamode1-test2",
				Tags: map[string]string{"wukong": "000"},
				EpId: "0",
			},
			IP: "192.168.20.53,100.93.4.203",
		},
	}, nil)
	patches.ApplyFunc(loadAgentDimensions, func(instanceID string) {})
	defer patches.Reset()
	ecsInfo1 := ECSInfo{}
	_, filterMetrics := ecsInfo1.GetResourceInfo()
	assert.NotNil(t, filterMetrics)
}

func TestAGTECSInfo_GetResourceInfo(t *testing.T) {
	unit := "%"
	namespace := "AGT.ECS"
	metricInfosList := model.MetricInfoList{
		Unit:       unit,
		Namespace:  namespace,
		MetricName: "cpu_usage",
		Dimensions: []model.MetricsDimension{
			{
				Name:  "instance_id",
				Value: "9234ad9f-87a5-49a9-b6ec-11ecde1d943b",
			},
		},
	}
	patches := gomonkey.ApplyFuncReturn(getECSAGTMetrics, []model.MetricInfoList{metricInfosList})
	defer patches.Reset()
	agtEcsInfo1 := AGTECSInfo{}
	_, filterMetrics := agtEcsInfo1.GetResourceInfo()
	assert.NotNil(t, filterMetrics)
}
