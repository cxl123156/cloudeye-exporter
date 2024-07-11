package logs

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitLog(t *testing.T) {
	successFlag := true
	testCases := []struct {
		name    string
		patches func() *gomonkey.Patches
		expect  func(t *testing.T)
	}{
		{
			"error_case",
			func() *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				confLoader := &ConfLoader{}
				patches.ApplyMethod(confLoader, "LoadFile", func(c *ConfLoader, fPath string, cfg interface{}) error {
					successFlag = false
					return errors.New("path error")
				})
				patches.ApplyFuncReturn(newYamlLoader, confLoader)
				return patches
			},
			func(t *testing.T) {
				assert.True(t, successFlag)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			patches := testCase.patches()
			defer patches.Reset()
			InitLog("/logs.yml")
			testCase.expect(t)
		})
	}
}
