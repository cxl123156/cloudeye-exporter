package collector

import (
	"fmt"
	"time"

	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	eps "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/region"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var epsInfo = &EpsInfo{
	EpDetails: make([]model.EpDetail, 0),
}

type EpsInfo struct {
	EpDetails []model.EpDetail
	TTL       int64
}

const HelpInfo = `# HELP huaweicloud_epinfo huaweicloud_epinfo
# TYPE huaweicloud_epinfo gauge
`

func getEPSClient() *eps.EpsClient {
	return eps.NewEpsClient(getEPSClientBuilder().Build())
}

func getEPSClientBuilder() *http_client.HcHttpClientBuilder {
	builder := eps.EpsClientBuilder().WithCredential(global.NewCredentialsBuilder().
		WithAk(conf.AccessKey).WithSk(conf.SecretKey).
		WithDomainId(conf.DomainID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify))
	if endpoint, ok := endpointConfig["eps"]; ok {
		builder.WithEndpoint(endpoint)
	} else {
		builder.WithRegion(region.ValueOf("cn-north-4"))
	}
	return builder
}

func GetEPSInfo() (string, error) {
	result := HelpInfo
	epsInfo, err := listEps()
	if err != nil {
		return result, err
	}
	for _, detail := range epsInfo {
		result += fmt.Sprintf("%s_epinfo{epId=\"%s\",epName=\"%s\"} 1\n", CloudConf.Global.Prefix, detail.Id, detail.Name)
	}
	return result, nil
}

func listEps() ([]model.EpDetail, error) {
	if epsInfo != nil && time.Now().Unix() < epsInfo.TTL {
		return epsInfo.EpDetails, nil
	}

	limit := int32(1000)
	Offset := int32(0)
	req := &model.ListEnterpriseProjectRequest{
		Limit:  &limit,
		Offset: &Offset,
	}

	client := getEPSClient()
	var resources []model.EpDetail
	for {
		response, err := client.ListEnterpriseProject(req)
		if err != nil {
			return resources, err
		}
		resources = append(resources, *response.EnterpriseProjects...)
		if len(*response.EnterpriseProjects) == 0 {
			break
		}
		if len(resources) > MaxEpsCount {
			logs.Logger.Errorf("eps not allowed to exceed 10000")
			break
		}
		*req.Offset += limit
	}
	epsInfo.EpDetails = resources
	epsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	return epsInfo.EpDetails, nil
}
