package collector

import (
	"testing"

	nosqlmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestNoSQLInfo_GetResourceInfo(t *testing.T) {
	instance := []nosqlmodel.ListInstancesResult{
		{
			Id:   "1111",
			Name: "test_xxx",
			Datastore: &nosqlmodel.ListInstancesDatastoreResult{
				Type: "redis",
			},
			Groups: []nosqlmodel.ListInstancesGroupResult{
				{
					Nodes: []nosqlmodel.ListInstancesNodeResult{
						{
							Id:   "test_node111",
							Name: "test_node1",
						},
						{
							Id:   "test_node222",
							Name: "test_node2",
						},
					},
				},
			},
		},
	}
	sysConfig := map[string][]string{
		"redis_cluster_id":               {"redis667_cluster_qps"},
		"redis_cluster_id,redis_node_id": {"gemini001_cpu_usage"},
		"mongodb_cluster_id":             {"mongodb001_command_ps"},
	}
	patches := getPatches()
	patches = patches.ApplyFuncReturn(getAllNoSQLInstances, instance, nil)
	patches = patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	defer patches.Reset()
	logs.InitLog("")
	noSQLInfo1 := NoSQLInfo{}
	_, filteredMetricInfos := noSQLInfo1.GetResourceInfo()
	assert.NotNil(t, filteredMetricInfos)
}
