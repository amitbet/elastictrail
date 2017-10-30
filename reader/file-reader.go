package reader

import "elastictrail/common"

type FileReader struct {
	common.ConsumerDictionary // the consumer registry
	FileName                  string
}

func (fr *FileReader) Read() {

}
