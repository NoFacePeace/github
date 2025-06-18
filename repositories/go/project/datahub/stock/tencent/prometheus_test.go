package tencent

import (
	"testing"

	"github.com/NoFacePeace/github/repositories/go/utils/datetime"
	"github.com/stretchr/testify/suite"
)

type PrometheusTestSuite struct {
	suite.Suite
	Prometheus *Prometheus
}

func (suite *PrometheusTestSuite) SetupSuite() {
	suite.Prometheus = NewPrometheus(&Config{
		Address: "http://127.0.0.1:9090/api/v1/write",
	})
}

func TestPrometheusTestSuite(t *testing.T) {
	suite.Run(t, new(PrometheusTestSuite))
}

func (suite *PrometheusTestSuite) TestWrite() {
	err := suite.Prometheus.Write([]Point{
		{
			Metric: "test11",
			Labels: map[string]string{
				"env": "test",
				"hh":  "test",
			},
			Time: datetime.Yesterday(),
			// Time:  time.Now(),
			Value: 4,
		},
	})
	suite.NoError(err)
}
