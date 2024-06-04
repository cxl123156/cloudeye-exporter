package collector

import (
	"errors"
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	rmsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestBuildElasticPoolWhenMapIsEmpty(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	sysConfigMap := getMetricConfigMap("SYS.DLI")
	buildElasticPool(sysConfigMap, &filterMetrics, resourceInfos)
	assert.Nil(t, sysConfigMap)
}

func TestBuildElasticPoolWhenGetError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	sysConfigMap := map[string][]string{}

	metricArr := make([]string, 0, 0)
	metricArr = append(metricArr, "elastic_resource_pool_cpu_usage")
	sysConfigMap["elastic_resource_pool_id"] = metricArr
	buildElasticPool(sysConfigMap, &filterMetrics, resourceInfos)
	assert.Equal(t, 0, len(filterMetrics))
}

func TestBuildElasticPool(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	sysConfigMap := map[string][]string{}
	metricArr := make([]string, 0, 0)
	metricArr = append(metricArr, "elastic_resource_pool_cpu_usage")
	sysConfigMap["elastic_resource_pool_id"] = metricArr

	respPage1 := ElasticPoolResponse{
		HttpStatusCode: 200,
		ElasticPools: []ElasticPool{
			{ID: 442, EpID: "0", PoolName: "Max"},
		},
	}
	respPage2 := ElasticPoolResponse{
		HttpStatusCode: 200,
		ElasticPools:   []ElasticPool{},
	}
	patches.ApplyMethodFunc(getHcClient(getEndpoint("dli", "v3")), "Sync", func(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		request, ok := req.(*ListFlinkJobsRequest)
		if !ok {
			return nil, errors.New("test error")
		}
		if *request.Offset == 0 {
			return &respPage1, nil
		}
		return &respPage2, nil
	})

	buildElasticPool(sysConfigMap, &filterMetrics, resourceInfos)
	assert.Equal(t, 1, len(filterMetrics))
}

func TestGetQueuesFromRMS(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()
	id1 := "123123123"
	name1 := "test_name"
	response := []rmsModel.ResourceEntity{
		{
			Id:   &id1,
			Name: &name1,
			Properties: map[string]interface{}{
				"queue_id": "2222",
			},
		},
	}
	patches.ApplyFuncReturn(listResources, response, nil)
	resources, err := getQueuesFromRMS()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resources))
}
