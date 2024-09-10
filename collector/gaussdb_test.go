package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestGaussdbGetResourceInfo(t *testing.T) {
	metricConf = map[string]MetricConf{
		"SYS.GAUSSDB": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"gaussdb_mysql_instance_id,gaussdb_mysql_node_id": {"gaussdb_mysql114_innodb_bufpool_read_ahead"}, "gaussdb_mysql_ha_id,gaussdb_mysql_ha_node_id": {"rds001_cpu_util"},
			},
		},
		"SYS.DBPROXY": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"dbproxy_instance_id,dbproxy_node_id": {"l4_out_bps_usage"},
			},
		},
	}
	gaussdbInfo.LabelInfo = nil
	instances := mockRmsResource()

	nodes := mockRmsResource()
	nodes[0].Properties = map[string]interface{}{
		"dimensions": []model.MetricsDimension{
			{Name: "gaussdb_mysql_instance_id", Value: "0001-0001-000001"},
			{Name: "gaussdb_mysql_node_id", Value: "node-0001-0001-0000001"},
		},
		"instanceId": "0001-0001-000001",
	}

	proxyNodes := mockRmsResource()
	proxyNodes[0].Properties = map[string]interface{}{
		"masterInstanceId": "0001-0001-000001",
		"role":             "master",
		"ipAddress":        "192.xxx.xxx.xxx",
		"proxyId":          "0002-0002-0000002",
		"engineName":       "innodb",
	}

	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{nodes, nil},
		},
		{
			Values: gomonkey.Params{instances, nil},
		},
		{
			Values: gomonkey.Params{proxyNodes, nil},
		},
	}

	patches := gomonkey.ApplyFuncSeq(listResources, outputs)
	defer patches.Reset()

	var gaussdbgetter GAUSSDBInfo
	labels, metrics := gaussdbgetter.GetResourceInfo()
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, 2, len(metrics))
}

func TestGaussdbGetHaResourceInfo(t *testing.T) {
	metricConf = map[string]MetricConf{
		"SYS.GAUSSDB": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"gaussdb_mysql_instance_id,gaussdb_mysql_node_id": {"gaussdb_mysql114_innodb_bufpool_read_ahead"}, "gaussdb_mysql_ha_id,gaussdb_mysql_ha_node_id": {"rds001_cpu_util"},
			},
		},
	}
	gaussdbInfo.LabelInfo = nil
	nodes := mockRmsResource()
	nodes[0].Properties = map[string]interface{}{
		"dimensions": []model.MetricsDimension{
			{Name: "gaussdb_mysql_ha_id", Value: "0001-0001-000001"},
			{Name: "gaussdb_mysql_ha_node_id", Value: "node-0001-0001-0000001"},
		},
		"instanceId": "0001-0001-000001",
	}
	patches := gomonkey.ApplyFuncReturn(listResources, nodes, nil)
	defer patches.Reset()

	var gaussdbgetter GAUSSDBInfo
	labels, metrics := gaussdbgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
