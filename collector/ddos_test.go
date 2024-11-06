package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestDdosGetResourceInfo(t *testing.T) {
	metricConf = map[string]MetricConf{
		"SYS.DDOS": {
			DimMetricName: map[string][]string{"instance_id": {"instance_drop_rate"},
				"package": {"instance_drop_rate"}, "package,ip": {"ip_drop_rate"},
			},
		},
	}
	patches := gomonkey.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()
	ddosClient := getDDOSClient()
	patches.ApplyMethodFunc(ddosClient, "ListProtectedIp", func(req *model.ListProtectedIpRequest) (*model.
		ListProtectedIpResponse, error) {
		var totalCount int32 = 1
		items := []model.ProtectedIpResponse{
			{
				Id:          "test-protected-ip-id",
				PackageId:   "0001-0001-000001",
				PackageName: "test-package-name",
			},
		}
		return &model.ListProtectedIpResponse{
			Total:          &totalCount,
			Items:          &items,
			HttpStatusCode: 200,
		}, nil
	})

	patches.ApplyMethodFunc(ddosClient, "ListPackage", func(req *model.ListPackageRequest) (*model.
		ListPackageResponse, error) {
		var totalCount int32 = 1
		items := []model.PackageResponse{
			{
				PackageId:   "0001-0001-000003",
				PackageName: "test-package-name",
			},
		}
		return &model.ListPackageResponse{
			Total:          &totalCount,
			Items:          &items,
			HttpStatusCode: 200,
		}, nil
	})

	var ddosgetter DDOSInfo
	labels, metrics := ddosgetter.GetResourceInfo()
	assert.Equal(t, 3, len(labels))
	assert.Equal(t, 3, len(metrics))
}
