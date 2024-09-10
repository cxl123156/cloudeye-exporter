package collector

import (
	"errors"
	"strings"

	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	v1 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/region"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func getRMSClient() *v1.RmsClient {
	return v1.NewRmsClient(getRMSClientBuilder().Build())
}

func getRMSClientBuilder() *http_client.HcHttpClientBuilder {
	builder := v1.RmsClientBuilder().WithCredential(global.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithDomainId(conf.DomainID).Build()).WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify))
	if endpoint, ok := endpointConfig["rms"]; ok {
		builder.WithEndpoint(endpoint)
	} else {
		builder.WithRegion(region.ValueOf("cn-north-4"))
	}
	return builder
}

func listResources(provider, resourceType string, optionalRegionID ...string) ([]model.ResourceEntity, error) {
	regionID := conf.Region
	if len(optionalRegionID) > 0 {
		regionID = optionalRegionID[0]
	}
	limit := int32(200)
	var resources []model.ResourceEntity
	req := &model.ListResourcesRequest{
		Provider: provider,
		Type:     resourceType,
		RegionId: &regionID,
		Limit:    &limit,
	}
	if CloudConf.Global.EpIds != "" {
		epIdArr := strings.Split(CloudConf.Global.EpIds, ",")
		for _, epID := range epIdArr {
			req.EpId = &epID
			resourceByEpID, err := getResourcesFromRMS(req)
			if err != nil {
				logs.Logger.Errorf("Get resources from rms by epID failed, epID is %s, error: %s", epID, err.Error())
				return nil, errors.New("get resources from rms by epID failed")
			}
			resources = append(resources, resourceByEpID...)
		}
		return resources, nil
	} else {
		return getResourcesFromRMS(req)
	}
}

func getResourcesFromRMS(req *model.ListResourcesRequest) ([]model.ResourceEntity, error) {
	var resources []model.ResourceEntity
	for {
		response, err := getRMSClient().ListResources(req)
		if err != nil {
			return resources, err
		}
		resources = append(resources, *response.Resources...)
		if response.PageInfo.NextMarker == nil {
			break
		}
		req.Marker = response.PageInfo.NextMarker
	}

	if len(resources) == 0 {
		logs.Logger.Infof("Get empty resource list from rms")
	}

	return resources, nil
}
