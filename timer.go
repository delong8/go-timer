package timer

import (
	"errors"
	"fmt"
	"log"
)

var (
	logger = log.New(log.Writer(), "[go-timer] ", log.LstdFlags)
	daily  = NewDaily()
)

func Init() {
	logger.Println("init timer")
	daily.Start()
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
func RegisteInterval(name string, minute int, fn func() string) error {
	if minute < 1 {
		return errors.New("minute must be equal or over then 1")
	}
	for i := 0; i < 1440; i += minute {
		err := daily.RegisteTask(fmt.Sprintf("%s%s%d", name, INTERVAL_SEPERATER, i), fn, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// cancel timer
func Cancel(name string) []CanceledTask {
	return daily.CancelTask(name)
}
