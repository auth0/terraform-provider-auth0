// Package wait contains helper functions to enable polling the api server.
package wait

import (
	"fmt"
	"time"
)

// Until takes a function which is called every pollingMillis milliseconds until
// it returns true, it has tried polling more than pollCount times, or there is an error.
func Until(pollingMillis, pollCount int, pollFunc func() (bool, error)) error {
	if pollingMillis < 0 {
		return fmt.Errorf("pollingMillis must not be less than zero")
	} else if pollCount < 0 {
		return fmt.Errorf("pollingCount must not be less than zero")
	}

	for i := 0; i < pollCount; i++ {
		if condition, err := pollFunc(); err != nil {
			return err
		} else if condition {
			return nil
		}
		time.Sleep(time.Duration(pollingMillis) * time.Millisecond)
	}
	return fmt.Errorf("timeout after %d milliseconds", pollingMillis*pollCount)
}
