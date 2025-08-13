package timer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DailyTaskQueue struct {
	started bool
	// 每分钟 1 个 tick，每天 1440 个 tick
	tick int
	// 日期计数，记录当前循环处于哪个日期
	date    string
	tasks   []DailyTask
	results []DailyTaskResult
}

type DailyTask struct {
	Running bool
	Fn      func() string
	Name    string
	// 在一个周期内，第多少个 tick 开始执行
	RunAtTick int
	// 上次执行是在哪个执行周期
	RunAtDate string
}

type DailyTaskResult struct {
	Name     string
	Message  string
	Error    error
	StartAt  time.Time
	EndAt    time.Time
	Manually bool // 是否手动触发
}

func parseTimeToTick(hm string) (tick int, err error) {
	t := strings.Split(hm, ":")
	if len(t) != 2 {
		err = fmt.Errorf("wrong format:%s", hm)
		return
	}
	h, err := strconv.ParseInt(t[0], 10, 8)
	if err != nil {
		return
	}
	m, err := strconv.ParseInt(t[1], 10, 8)
	if err != nil {
		return
	}
	if h < 0 || h > 23 {
		err = fmt.Errorf("hour must be in 0-23")
		return
	}
	if m < 0 || m > 59 {
		err = fmt.Errorf("minute must be in 0-59")
		return
	}
	return int(h*60 + m), nil
}

func (tq *DailyTaskQueue) RegisteTask(name string, fn func() string, tick int) error {
	if tick < 0 || tick > 1399 {
		return fmt.Errorf("tick must be in 0-1399, not is %d", tick)
	}
	for _, t := range tq.tasks {
		if t.Name == name {
			return fmt.Errorf("task already exists: %s", name)
		}
	}
	fmt.Println(333, name)
	tq.tasks = append(tq.tasks, DailyTask{
		Fn:        fn,
		Name:      name,
		RunAtTick: tick,
	})
	return nil
}

func (tq *DailyTaskQueue) Start() {
	if tq.started {
		return
	}
	tq.started = true
	go tq.looper()
}

func (tq *DailyTaskQueue) move() {
	now := time.Now()
	tq.tick = now.Hour()*60 + now.Minute()
	tq.date = now.Format("2006-01-02")
	fmt.Println(tq.tick, tq.date)
}

func (tq *DailyTaskQueue) looper() {
	for {
		tq.move()
		for _, task := range tq.tasks {
			fmt.Println(222, task.Name)
			if tq.shouldRun(task) {
				tq.caller(task, false)
			}
		}
		time.Sleep(time.Minute)
	}
}

func (tq *DailyTaskQueue) shouldRun(task DailyTask) bool {
	fmt.Println("should run", task.RunAtTick, task.RunAtDate, tq.date, tq.tick)
	// 检查是否到了执行时间
	if tq.date == task.RunAtDate {
		return false
	}
	return tq.tick >= task.RunAtTick
}

func (tq *DailyTaskQueue) caller(t DailyTask, manually bool) {
	rst := DailyTaskResult{
		Name:     t.Name,
		StartAt:  time.Now(),
		Manually: manually,
	}

	defer func() {
		if err := recover(); err != nil {
			// log.Error(err)
		}
		t.Running = false
		tq.appendResult(rst)
	}()

	if t.Running {
		rst.Error = errors.New("caller is running")
		return
	}
	t.Running = true

	rst.Message = t.Fn()
	// log.Info(t.Name, rst.Message)
	rst.EndAt = time.Now()
}

func (tq *DailyTaskQueue) appendResult(rst DailyTaskResult) {
	tq.results = append(tq.results, rst)
	if len(tq.results) > 1000 {
		tq.results = tq.results[2:]
	}
}

// manually run a task
func (tq *DailyTaskQueue) RunTask(name string) error {
	for _, task := range tq.tasks {
		if task.Name == name {
			tq.caller(task, true)
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", name)
}

func (tq *DailyTaskQueue) Status() []DailyTask {
	return tq.tasks
}

func (tq *DailyTaskQueue) History() []DailyTaskResult {
	return tq.results
}
