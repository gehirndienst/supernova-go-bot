package fetch

import (
	"errors"

	"github.com/rs/zerolog"
)

type Fetchable interface {
	Set(APIKey string, logger *zerolog.Logger) error
	Fetch(qParams map[string]interface{}) (string, error)
}

type BaseFetcher struct {
	APIKey string
	logger *zerolog.Logger
}

func (bf *BaseFetcher) isSet() bool {
	return bf.APIKey != "" && bf.logger != nil
}

func (bf *BaseFetcher) Set(APIKey string, logger *zerolog.Logger) error {
	if logger == nil {
		return errors.New("logger is required")
	}
	if APIKey == "" {
		return errors.New("APIKey is required")
	}
	bf.APIKey = APIKey
	bf.logger = logger
	return nil
}
