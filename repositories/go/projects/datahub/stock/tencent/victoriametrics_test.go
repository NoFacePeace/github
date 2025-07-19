package tencent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type VictoriaMetricsTestSuite struct {
	suite.Suite
	VictoriaMetrics *VictoriaMetrics
}

func (suite *VictoriaMetricsTestSuite) SetupSuite() {
	suite.VictoriaMetrics = NewVictoriaMetrics("http://localhost:8428/api/v1/import")
}

func TestVictoriaMetricsTestSuite(t *testing.T) {
	suite.Run(t, new(VictoriaMetricsTestSuite))
}

func (suite *VictoriaMetricsTestSuite) TestWrite() {
	data := VictoriaMetricsData{
		Metric: map[string]string{
			"__name__": "test",
			"env":      "test",
		},
		Values:     []float64{2},
		Timestamps: []int64{time.Now().UnixMilli()},
	}
	err := suite.VictoriaMetrics.Write(data)
	suite.Require().Nil(err)
}
