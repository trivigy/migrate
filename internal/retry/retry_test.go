package retry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RetrySuite struct {
	suite.Suite
}

func (r *RetrySuite) TestRetry_Do() {
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	count := 0
	assert.Nil(r.T(), Do(ctx, 100*time.Millisecond, func() (bool, error) {
		count++
		if count >= 5 {
			return false, nil
		}
		return true, nil
	}))
	assert.True(r.T(), count >= 4)
}

func TestRouterSuite(t *testing.T) {
	suite.Run(t, new(RetrySuite))
}
