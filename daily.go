package timer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type dailyTaskQueue struct {
	started bool
	// 每分钟 1 个 tick，每天 1440 个 tick
	tick int
	// 日期计数，记录当前循环处于哪个日期
	date    string
	tasks   []*dailyTask
	results []dailyTaskResult
}

type dailyTask struct {
	Running bool
	Fn      func() string
	Name    string
	// 在一个周期内，第多少个 tick 开始执行
	RunAtTick int
	// 上次执行是在哪个执行周期
	RunAtDate string
	Canceled  bool
}

type dailyTaskResult struct {
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

func (tq *dailyTaskQueue) RegisteTask(name string, fn func() string, tick int) error {
	if tick < 0 || tick > 1399 {
		return fmt.Errorf("tick must be in 0-1399, not is %d", tick)
	}
	for _, t := range tq.tasks {
		if t.Name == name {
			return fmt.Errorf("task already exists: %s", name)
		}
	}
	// if tick less than now, then mark it has been executed
	runAtDate := ""
	if tick <= tq.tick {
		runAtDate = tq.date
	}
	tq.tasks = append(tq.tasks, &dailyTask{
		Fn:        fn,
		Name:      name,
		RunAtTick: tick,
		RunAtDate: runAtDate,
	})
	return nil
}

type CanceledTask struct {
	Name string
	Tick int
}

func (tq *dailyTaskQueue) CancelTask(name string) []CanceledTask {
	canceledTask := make([]CanceledTask, 0)
	for _, t := range tq.tasks {
		if t.Name != name {
			canceledTask = append(canceledTask, CanceledTask{
				Name: t.Name,
				Tick: t.RunAtTick,
			})
			t.Canceled = true
		}
	}
	return canceledTask
}

func (tq *dailyTaskQueue) Start() {
	if tq.started {
		logger.Println("daily task queue already started")
		return
	}
	tq.started = true

	// Calculate delay to start at the beginning of the next minute
	now := time.Now()
	delay := time.Duration(60-now.Second()) * time.Second
	logger.Printf("daily task queue will start in %v seconds", delay.Seconds())

	go func() {
		time.Sleep(delay)
		tq.looper()
	}()

	logger.Println("daily task queue started")
}

func (tq *dailyTaskQueue) move() {
	now := time.Now()
	tq.tick = now.Hour()*60 + now.Minute()
	tq.date = now.Format("2006-01-02")
}

func (tq *dailyTaskQueue) looper() {
	for {
		var tasks []*dailyTask
		tq.move()
		// remove canceled tasks
		for _, t := range tq.tasks {
			if t.Canceled {
				continue
			}
			tasks = append(tasks, t)
		}
		// update task list
		tq.tasks = tasks

		for _, task := range tq.tasks {
			if tq.shouldRun(*task) {
				tq.caller(*task, false)
			}
		}
		time.Sleep(time.Minute)
	}
}

func (tq *dailyTaskQueue) shouldRun(task dailyTask) bool {
	// 检查是否到了执行时间
	if tq.date == task.RunAtDate {
		return false
	}
	// check if the task is been canceled
	if task.Canceled {
		return false
	}
	return tq.tick >= task.RunAtTick
}

func (tq *dailyTaskQueue) caller(t dailyTask, manually bool) {
	rst := dailyTaskResult{
		Name:     t.Name,
		StartAt:  time.Now(),
		Manually: manually,
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Println("Error:", err)
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
	rst.EndAt = time.Now()
}

func (tq *dailyTaskQueue) appendResult(rst dailyTaskResult) {
	tq.results = append(tq.results, rst)
	if len(tq.results) > 1000 {
		tq.results = tq.results[2:]
	}
}

// manually run a task
func (tq *dailyTaskQueue) RunTask(name string) error {
	for _, task := range tq.tasks {
		if task.Name == name {
			tq.caller(*task, true)
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", name)
}

func (tq *dailyTaskQueue) Status() []dailyTask {
	ts := make([]dailyTask, 0)
	for _, t := range tq.tasks {
		ts = append(ts, *t)
	}
	return ts
}

func (tq *dailyTaskQueue) History() []dailyTaskResult {
	return tq.results
}
