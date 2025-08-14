package timer

import (
	"errors"
	// "github.com/delong8/rlog"
)

var (
	// log     = rlog.New("timer")
	daily = dailyTaskQueue{}
)

func Init() {
	daily.Start()

	// 注册一个默认的 "try" 任务，每隔 10 分钟执行一次
	RegisteInterval("try", 10, func() string {
		return "try task executed"
	})
}

// 添加任务, 时间格式为 09:12，每天执行一次
func RegisteDaily(name string, time string, fn func() string) error {
	tick, err := parseTimeToTick(time)
	if err != nil {
		return err
	}
	return daily.RegisteTask(name, fn, tick)
}

// 添加定时任务，每隔 xx 分钟执行一次
func RegisteInterval(name string, minutes int, fn func() string) error {
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
