package collector

import (
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestGetRmsClient(t *testing.T) {
	endpointConfig = map[string]string{
		"rms": "https://rms.myhuaweicloud.com",
	}
	client := getRMSClient()
	assert.NotNil(t, client)
}

func TestListResources(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()

	id1 := "123123123"
	name1 := "test_name"
	currentNum := int32(1)
	epID := "0"
	response := model.ListResourcesResponse{
		Resources: &[]model.ResourceEntity{
			{
				Id:   &id1,
				Name: &name1,
				EpId: &epID,
				Properties: map[string]interface{}{
					"queue_id": "2222",
				},
			},
		},
		PageInfo: &model.PageInfo{
			CurrentCount: &currentNum,
		},
	}
	endpointConfig = map[string]string{
		"rms": "https://rms.myhuaweicloud.com",
	}

	patches.ApplyMethodFunc(getRMSClient(), "ListResources", func(request *model.ListResourcesRequest) (*model.ListResourcesResponse, error) {
		return &response, nil
	})

	CloudConf.Global.EpIds = "0"
	resources, err := listResources("ecs", "cloudservers")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resources))
}

func TestListAllResourcesFromRMS(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()

	id1 := "123123123"
	name1 := "test_name"
	currentNum := int32(1)
	epID1 := "0"
	epID2 := "xxxxx"
	response := model.ListResourcesResponse{
		Resources: &[]model.ResourceEntity{
			{
				Id:   &id1,
				Name: &name1,
				EpId: &epID1,
				Properties: map[string]interface{}{
					"queue_id": "2222",
				},
			},
			{
				Id:   &id1,
				Name: &name1,
				EpId: &epID2,
				Properties: map[string]interface{}{
					"queue_id": "2222",
				},
			},
		},
		PageInfo: &model.PageInfo{
			CurrentCount: &currentNum,
		},
	}
	endpointConfig = map[string]string{
		"rms": "https://rms.myhuaweicloud.com",
	}

	patches.ApplyMethodFunc(getRMSClient(), "ListResources", func(request *model.ListResourcesRequest) (*model.ListResourcesResponse, error) {
		return &response, nil
	})

	resources, err := listResources("ecs", "cloudservers")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resources))
}
