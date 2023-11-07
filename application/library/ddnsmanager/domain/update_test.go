package domain

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	res := map[int]int{}
	for i := 1; i <= 10; i++ {
		res[i] = i
	}
	var errs []error
	wg := sync.WaitGroup{}
	chanErrors := make(chan error, len(res)+10)
	if true {
		for k := range res {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					time.Sleep(time.Second)
					fmt.Printf("[%d] %s\n", k, time.Now().Format(time.DateTime))
				}
				chanErrors <- fmt.Errorf(`test-error: %d`, k)
			}(k)
		}
	}
	wg.Wait()
	close(chanErrors)
	for chanErr := range chanErrors {
		if chanErr == nil {
			continue
		}
		t.Error(chanErr.Error())
		errs = append(errs, chanErr)
	}
}
