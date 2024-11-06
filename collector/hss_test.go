package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestGetHSSResourceInfo(t *testing.T) {
	metricConf = make(map[string]MetricConf)
	metricConf["SYS.HSS"] = MetricConf{
		Resource: "service",
		DimMetricName: map[string][]string{
			"hss_enterprise_project_id": {"host_num", "unprotected_host_num", "risky_host_num", "uninstalled_or_offline_agent_num"},
		},
	}
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(listEps, []model.EpDetail{
		{
			Id:          "0",
			Name:        "default",
			Description: "默认企业项目",
			Status:      1,
		},
	}, nil)
	defer patches.Reset()
	var hssInfo HSSInfo
	info, lists := hssInfo.GetResourceInfo()
	assert.NotNil(t, info)
	assert.NotNil(t, lists)
}
