package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestVPCEPInfo_GetResourceInfo(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(getVpcEpEndpoints, []VpcEpEndpoint{}, nil)
	defer patches.Reset()
	vpcEpInfo1 := VPCEPInfo{}
	_, filteredMetricInfo := vpcEpInfo1.GetResourceInfo()
	assert.NotNil(t, filteredMetricInfo)
}
