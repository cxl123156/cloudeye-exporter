package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	eps "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestGetEPSClient(t *testing.T) {
	endpointConfig = map[string]string{
		"eps": "https://eps.myhuaweicloud.com",
	}
	epsClient := getEPSClient()
	assert.NotNil(t, epsClient)
}

func TestGetEPSInfo(t *testing.T) {
	endpointConfig = map[string]string{
		"eps": "https://eps.myhuaweicloud.com",
	}
	client := eps.NewEpsClient(getEPSClientBuilder().Build())
	patches := gomonkey.NewPatches()
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListEnterpriseProjectResponse{
				EnterpriseProjects: &[]model.EpDetail{
					{
						Id:          "0",
						Name:        "default",
						Description: "默认企业项目",
						Status:      1,
					},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListEnterpriseProjectResponse{
				EnterpriseProjects: &[]model.EpDetail{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(client, "ListEnterpriseProject", outputs)
	patches.ApplyFuncReturn(eps.NewEpsClient, client)
	defer patches.Reset()
	_, err := GetEPSInfo()
	assert.Nil(t, err)
}
