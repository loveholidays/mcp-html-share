package coverageprobe_test

import (
	"testing"

	"github.com/loveholidays/mcp-html-share/internal/coverageprobe"
	"github.com/stretchr/testify/suite"
)

type ProbeSuite struct {
	suite.Suite
}

func TestProbeSuite(t *testing.T) {
	suite.Run(t, new(ProbeSuite))
}

func (s *ProbeSuite) TestNegativeValue() {
	s.Zero(coverageprobe.Increment(-1))
}

func (s *ProbeSuite) TestZeroValue() {
	s.Equal(1, coverageprobe.Increment(0))
}

func (s *ProbeSuite) TestPositiveValue() {
	s.Equal(3, coverageprobe.Increment(2))
}
