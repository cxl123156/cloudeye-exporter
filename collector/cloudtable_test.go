package collector

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cloudtable/v2/model"
	"github.com/stretchr/testify/assert"
)

func TestCloudTableInfo_GetResourceInfo(t *testing.T) {
	clusterID := "cls-0001"
	clusterName := "cluster1"
	respPage1 := model.ListClustersResponse{
		HttpStatusCode: 200,
		Clusters: &[]model.ClusterDetail{
			{ClusterId: &clusterID, ClusterName: &clusterName},
		},
	}
	respPage2 := model.ListClustersResponse{
		HttpStatusCode: 200,
		Clusters:       &[]model.ClusterDetail{},
	}
	cloudTableClient := getCloudTableClient()
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyMethodFunc(cloudTableClient.HcClient, "Sync", func(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		request, ok := req.(*model.ListClustersRequest)
		if !ok {
			return nil, errors.New("test error")
		}
		if *request.Offset == 0 {
			return &respPage1, nil
		}
		return &respPage2, nil
	})
	patches.ApplyFuncReturn(listAllMetrics, []cesmodel.MetricInfoList{
		{
			Namespace:  "SYS.CloudTable",
			MetricName: "cmdProcessCPU",
			Dimensions: []cesmodel.MetricsDimension{
				{
					Name:  "cluster_id",
					Value: "cls-0001",
				},
				{
					Name:  "instance_name",
					Value: "server-1-1",
				},
			},
		},
	}, nil)
	var getter CloudTableInfo
	resourceInfos, filteredMetrics := getter.GetResourceInfo()
	assert.NotNil(t, resourceInfos)
	assert.NotNil(t, filteredMetrics)
}
