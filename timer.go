package timer

import (
	"errors"
	// "github.com/delong8/rlog"
)

var (
	// log     = rlog.New("timer")
	daily = DailyTaskQueue{}
)

func Init() {
	daily.Start()
}

// 添加任务, 时间格式为 09:12，每天执行一次
func RegisteDaily(name string, fn func() string, time string) error {
	tick, err := parseTimeToTick(time)
	if err != nil {
		return err
	}
	return daily.RegisteTask(name, fn, tick)
}

// 添加定时任务，每隔 xx 分钟执行一次
func RegisteInterval(name string, fn func() string, minutes int) error {
	if minutes < 1 {
		return errors.New("minutes must be over then 1")
	}
	for i := 0; i < 1440; i += minutes {
		err := daily.RegisteTask(name, fn, i)
		if err != nil {
			return err
		}
	}
	return nil
}
