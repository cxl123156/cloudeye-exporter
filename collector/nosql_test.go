package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	nosqlmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3/model"
	"github.com/stretchr/testify/assert"
)

func TestNoSQLInfo_GetResourceInfo(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(getAllNoSQLInstances, []nosqlmodel.ListInstancesResult{}, nil)
	defer patches.Reset()
	noSQLInfo1 := NoSQLInfo{}
	_, filteredMetricInfos := noSQLInfo1.GetResourceInfo()
	assert.NotNil(t, filteredMetricInfos)
}
