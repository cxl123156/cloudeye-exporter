package collector

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestWAFInfo_GetResourceInfo(t *testing.T) {
	tests := []struct {
		name     string
		wantNil  bool
		want1Nil bool
	}{
		{
			"getAllWafInstancesFromRMSErr",
			true,
			true,
		},
		{
			"normal",
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var patches *gomonkey.Patches
			if tt.name == "normal" {
				sysConfig := map[string][]string{"waf_instance_id": {"attacks"}}
				patches = gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
				patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
				wafClient := getWAFClient()
				ID := "1"
				InstanceName := "AA"
				resp := &model.ListInstanceResponse{
					HttpStatusCode: 200,
					Items: &[]model.ListInstance{
						{
							Id:           &ID,
							InstanceName: &InstanceName,
						},
					},
				}
				patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, nil)
			}
			if tt.name == "getAllWafInstancesFromRMSErr" {
				patches = getPatches()

				sysConfig := map[string][]string{"waf_instance_id": {"attacks"}}
				patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)
				patches.ApplyFuncReturn(listResources, mockRmsResource(), errors.New(""))

				logs.InitLog("")
			}
			getter := WAFInfo{}
			got, got1 := getter.GetResourceInfo()
			assert.Equalf(t, tt.wantNil, got == nil, "GetResourceInfo()")
			assert.Equalf(t, tt.want1Nil, got1 == nil, "GetResourceInfo()")
			getter.resetResourceInfo()
			if patches != nil {
				patches.Reset()
			}
		})
	}
}

func TestWafGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"waf_instance_id": {"attacks"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	defer patches.Reset()
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	wafClient := getWAFClient()
	ID := "1"
	InstanceName := "AA"
	resp := &model.ListInstanceResponse{
		HttpStatusCode: 200,
		Items: &[]model.ListInstance{
			{
				Id:           &ID,
				InstanceName: &InstanceName,
			},
		},
	}
	patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, nil)

	var wafgetter WAFInfo
	labels, metrics := wafgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
	wafgetter.resetResourceInfo()
}

func TestWafGetResourceInfo_getAllWafInstancesFromRMSErr(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	sysConfig := map[string][]string{"waf_instance_id": {"attacks"}}
	patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), errors.New(""))

	logs.InitLog("")
	var wafGetter WAFInfo
	labels, metrics := wafGetter.GetResourceInfo()
	assert.Nil(t, labels)
	assert.Nil(t, metrics)
	wafGetter.resetResourceInfo()
}

func TestWafGetResourceInfo_getAllPremiumWafInstancesNormal(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	sysConfig := map[string][]string{
		"waf_instance_id": {"attacks"},
		"instance_id":     {"cpu_util"},
	}
	patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	wafClient := getWAFClient()
	ID := "1"
	InstanceName := "AA"
	resp := &model.ListInstanceResponse{
		HttpStatusCode: 200,
		Items: &[]model.ListInstance{
			{
				Id:           &ID,
				InstanceName: &InstanceName,
			},
		},
	}
	patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, nil)
	logs.InitLog("")
	var wafGetter WAFInfo
	labels, metrics := wafGetter.GetResourceInfo()
	assert.Equal(t, 2, len(labels))
	assert.Equal(t, 2, len(metrics))
	wafGetter.resetResourceInfo()
}

func TestWafGetResourceInfo_getAllPremiumWafInstancesErr(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	sysConfig := map[string][]string{
		"waf_instance_id": {"attacks"},
		"instance_id":     {"cpu_util"},
	}
	patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	wafClient := getWAFClient()
	ID := "1"
	InstanceName := "AA"
	resp := &model.ListInstanceResponse{
		HttpStatusCode: 200,
		Items: &[]model.ListInstance{
			{
				Id:           &ID,
				InstanceName: &InstanceName,
			},
		},
	}
	patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, errors.New(""))

	logs.InitLog("")
	var wafGetter WAFInfo
	labels, metrics := wafGetter.GetResourceInfo()
	assert.Nil(t, labels)
	assert.Nil(t, metrics)
	wafGetter.resetResourceInfo()
}

func Test_getWAFClient(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{
			"normal",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.wantNil, getWAFClient() == nil, "getWAFClient()")
		})
	}
}

func Test_getAllPremiumWafInstances(t *testing.T) {
	tests := []struct {
		name    string
		wantNil bool
	}{
		{
			"normal",
			false,
		},
		{
			"HttpStatusCode ERR",
			true,
		},
		{
			"Request ERR",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := getPatches()
			logs.InitLog("")
			if tt.name == "normal" {
				wafClient := getWAFClient()
				resp := &model.ListInstanceResponse{
					HttpStatusCode: 200,
					Items:          &[]model.ListInstance{},
				}
				patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, nil)
			}
			if tt.name == "HttpStatusCode ERR" {
				wafClient := getWAFClient()
				resp := &model.ListInstanceResponse{
					HttpStatusCode: 404,
					Items:          &[]model.ListInstance{},
				}
				patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, nil)
			}
			if tt.name == "Request ERR" {
				wafClient := getWAFClient()
				resp := &model.ListInstanceResponse{
					HttpStatusCode: 200,
					Items:          &[]model.ListInstance{},
				}
				patches = gomonkey.ApplyMethodReturn(wafClient, "ListInstance", resp, errors.New("aa"))
			}
			assert.Equalf(t, tt.wantNil, getAllPremiumWafInstances() == nil, "getAllPremiumWafInstances()")
			if patches != nil {
				patches.Reset()
			}
		})
	}
}
