package reader

import (
	"elastictrail/common"
	"time"
)

type MockReader struct {
	Lines                     []string
	common.ConsumerDictionary // the consumer registry
}

func (r *MockReader) Start() {
	for _, l := range r.Lines {
		r.ConsumerDictionary.Distribute(&common.SimpleLine{MyMessage: l, Fields: map[string]string{"message": l}})
	}
	time.Sleep(time.Millisecond * 300)
	r.ConsumerDictionary.DistributeBatchDone()
}
