package ddnsretry

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/admpub/log"
)

var ErrCanceledRetry = errors.New(`canceled retry:`)
var RetrtDuration atomic.Int32 // 最大间隔秒数

func init() {
	RetrtDuration.Store(300)
}

func Retry(ctx context.Context, fn func(ctx context.Context) error, step ...int) (err error) {
	var _step int
	if len(step) > 0 {
		_step = step[0]
	}
	if _step <= 0 {
		_step = 10
	}
	tick := time.NewTicker(time.Second)
	retryCount := _step
	var lastTime time.Time
	startTime := time.Now()
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf(`%w: %v`, ErrCanceledRetry, err)
		case tm := <-tick.C:
			d := time.Second * time.Duration(retryCount)
			if lastTime.IsZero() || tm.Sub(lastTime) >= d {
				lastTime = tm
				err = fn(ctx)
				if err == nil {
					return
				}
				log.Errorf(`%v (Wait to try again later)`, err)
				retryCount += _step
			}
			if tm.Sub(startTime) >= time.Duration(RetrtDuration.Load())*time.Second {
				return
			}
		}
	}
}
