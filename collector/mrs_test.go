package collector

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestMRSInfo_GetResourceInfo(t *testing.T) {
	now := time.Now().Add(-time.Minute).Unix()
	mrsInfo1 := MRSInfo{}
	patches := gomonkey.ApplyFuncReturn(getMRSResourceAndMetrics, map[string]labelInfo{}, []model.MetricInfoList{})
	defer patches.Reset()
	mrsInfo1.GetResourceInfo()
	assert.NotNil(t, mrsInfo.LabelInfo)

	mrsInfo.TTL = now
	mrsInfo1.GetResourceInfo()
	assert.NotNil(t, mrsInfo.LabelInfo)
}
