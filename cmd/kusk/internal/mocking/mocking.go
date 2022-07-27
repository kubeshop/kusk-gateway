package mocking

import (
	"io"
)

const openApiMockingConfigStr = `application:
  debug: false
  log_format: json
  log_level: warn

generation:
  suppress_errors: false
  use_examples: 'if_present'
`

func WriteMockingConfig(w io.Writer) error {
	_, err := w.Write([]byte(openApiMockingConfigStr))
	return err
}
