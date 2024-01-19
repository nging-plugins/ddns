package boot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/admpub/confl"
	"github.com/admpub/log"
	nconfig "github.com/admpub/nging/v5/application/library/config"
	"github.com/admpub/nging/v5/application/library/config/startup"
	syncOnce "github.com/admpub/once"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/config"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/domain"
	"github.com/nging-plugins/ddnsmanager/application/library/ddnsretry"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"
)

var (
	dflt            = config.New()
	domains         *domain.Domains
	once            syncOnce.Once
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	defaultInterval = 5 * time.Minute
	waitingDuration = 500 * time.Millisecond
	forceUpdateSig  = make(chan struct{})
	forceUpdateErr  = make(chan error)
	ErrInitFail     = errors.New(`ddns boot failed`)
)

func IsRunning() bool {
	return cancel != nil
}

func Config() *config.Config {
	mutex.RLock()
	c := *dflt
	mutex.RUnlock()
	return &c
}

func init() {
	startup.OnAfter(`web.installed`, start)
}

func parseConfig() error {
	saveFile := filepath.Join(echo.Wd(), `config/ddns.yaml`)
	if !com.FileExists(saveFile) {
		return fmt.Errorf(`%w: %s`, os.ErrNotExist, saveFile)
	}
	_, err := confl.DecodeFile(saveFile, dflt)
	if err != nil {
		err = fmt.Errorf(`%s: %w`, saveFile, err)
	}
	return err
}

func start() {
	err := parseConfig()
	if err != nil {
		log.Error(err.Error())
		return
	}
	if dflt.Closed {
		return
	}
	err = Run(context.Background())
	if err != nil {
		log.Error(err)
	}
}

func SetConfig(c *config.Config) error {
	saveFile := filepath.Join(nconfig.FromCLI().ConfDir(), `ddns.yaml`)
	b, err := confl.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(saveFile, b, os.ModePerm)
	if err != nil {
		return err
	}
	mutex.Lock()
	*dflt = *c
	mutex.Unlock()
	return nil
}

func Run(rootCtx context.Context, intervals ...time.Duration) (err error) {
	cfg := Config()
	if !cfg.IsValid() {
		log.Warn(`[DDNS] Exit task: The task does not meet the startup conditions`)
		return nil
	}
	ddnsretry.RetrtDuration.Store(int32(cfg.Interval.Seconds()))
	d := Domains()
	if d == nil {
		return ErrInitFail
	}

	mutex.Lock()
	if cancel != nil {
		cancel()
		cancel = nil
		time.Sleep(waitingDuration)
	}
	ctx, cancel = context.WithCancel(rootCtx)
	mutex.Unlock()

	err = d.Update(ctx, cfg, false)
	if err != nil {
		log.Error(`[DDNS] Exit task`)
		return err
	}

	up := func(c context.Context, force bool) error {
		d := Domains()
		if d == nil {
			mutex.Lock()
			if cancel != nil {
				cancel()
				cancel = nil
			}
			mutex.Unlock()
			err = ErrInitFail
			return err
		}
		log.Debug(`[DDNS] Checking network ip`)
		err := d.Update(c, Config(), force)
		if err != nil {
			log.Error(err)
		}
		return nil
	}
	go func() {
		interval := cfg.Interval
		if len(intervals) > 0 {
			interval = intervals[0]
		}
		if interval < time.Second {
			interval = defaultInterval
		}
		t := time.NewTicker(interval)
		defer t.Stop()
		log.Okay(`[DDNS] Starting task. Interval: `, interval.String())
		for {
			select {
			case <-ctx.Done():
				log.Warn(`[DDNS] Forced exit task`)
				return
			case <-t.C:
				if err := up(ctx, false); err != nil {
					log.Error(`[DDNS] Exit task. Error: `, err.Error())
					return
				}
			case <-forceUpdateSig:
				select {
				case forceUpdateErr <- up(ctx, true):
				default:
				}
			}
		}
	}()
	return err
}

func Domains() *domain.Domains {
	once.Do(initDomains)
	return domains
}

func ForceUpdate(eCtx echo.Context) error {
	cfg := Config()
	if cfg.Closed {
		return eCtx.NewError(code.DataNotFound, `任务没有开启`)
	}
	d := Domains()
	if d == nil {
		return eCtx.NewError(code.DataNotFound, `任务启动失败，请查看日志了解详情`)
	}
	//return d.Update(ctx, cfg, true)

	t := time.NewTimer(time.Second * 3)
	defer t.Stop()
	for {
		select {
		case forceUpdateSig <- struct{}{}:
			return <-forceUpdateErr
		case <-t.C:
			return context.DeadlineExceeded
		}
	}
}

func Reset(ctx context.Context) error {
	cfg := Config() // 含锁，小心使用
	mutex.Lock()
	once.Reset()
	if cancel != nil {
		cancel()
		cancel = nil
		time.Sleep(waitingDuration)
		if cfg.Closed {
			log.Warn(`[DDNS] Stopping task`)
		}
	}
	mutex.Unlock()
	if cfg.Closed {
		return nil
	}
	log.Warn(`[DDNS] Starting reboot task`)
	return Run(ctx)
}

func initDomains() {
	err := commit()
	if err != nil {
		log.Error(err)
	}
}

func commit() error {
	err := dflt.Commit()
	if err != nil {
		return err
	}
	domains, err = domain.ParseDomain(dflt)
	return err
}
