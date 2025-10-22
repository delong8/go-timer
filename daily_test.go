package timer

import (
	"errors"
	"testing"
	"time"
)

func TestDailyTaskQueue_CancelTask(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 注册几个测试任务
	tq.RegisteTask("task1", func() string { return "test1" }, 540)
	tq.RegisteTask("task2", func() string { return "test2" }, 600)
	tq.RegisteTask("interval@1", func() string { return "interval test" }, 720)

	// 取消单个任务
	canceled := tq.CancelTask("task1")
	if len(canceled) != 1 {
		t.Errorf("Expected 1 canceled task, but got %d", len(canceled))
	}

	if canceled[0].Name != "task1" {
		t.Errorf("Expected canceled task name 'task1', but got '%s'", canceled[0].Name)
	}

	// 验证任务已被标记为取消
	for _, task := range tq.tasks {
		if task.Name == "task1" && !task.Canceled {
			t.Error("Expected task1 to be marked as canceled")
		}
	}

	// 取消带前缀的任务
	canceled = tq.CancelTask("interval")
	if len(canceled) != 1 {
		t.Errorf("Expected 1 canceled task, but got %d", len(canceled))
	}

	if canceled[0].Name != "interval@1" {
		t.Errorf("Expected canceled task name 'interval@1', but got '%s'", canceled[0].Name)
	}
}

func TestDailyTaskQueue_RunTask(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 注册测试任务
	result := ""
	tq.RegisteTask("run_task", func() string {
		result = "task executed"
		return result
	}, 540)

	// 手动运行任务
	err := tq.RunTask("run_task")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// 验证任务被执行
	if result != "task executed" {
		t.Errorf("Expected result 'task executed', but got '%s'", result)
	}

	// 尝试运行不存在的任务
	err = tq.RunTask("nonexistent_task")
	if err == nil {
		t.Error("Expected error for nonexistent task, but got none")
	}
}

func TestDailyTaskQueue_Status(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 清空任务列表
	tq.tasks = nil

	// 注册测试任务
	tq.RegisteTask("status_task1", func() string { return "test1" }, 540)
	tq.RegisteTask("status_task2", func() string { return "test2" }, 600)

	// 获取状态
	status := tq.Status()
	if len(status) != 2 {
		t.Errorf("Expected 2 tasks in status, but got %d", len(status))
	}

	// 验证任务信息
	foundTask1 := false
	foundTask2 := false
	for _, task := range status {
		if task.Name == "status_task1" {
			foundTask1 = true
		}
		if task.Name == "status_task2" {
			foundTask2 = true
		}
	}

	if !foundTask1 {
		t.Error("Expected to find status_task1 in status")
	}

	if !foundTask2 {
		t.Error("Expected to find status_task2 in status")
	}
}

func TestDailyTaskQueue_History(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 清空历史记录
	tq.results = nil

	// 添加一些测试结果
	result1 := dailyTaskResult{
		Name:    "history_task1",
		Message: "result1",
		Error:   nil,
	}

	result2 := dailyTaskResult{
		Name:    "history_task2",
		Message: "result2",
		Error:   errors.New("test error"),
	}

	tq.appendResult(result1)
	tq.appendResult(result2)

	// 获取历史记录
	history := tq.History()
	if len(history) != 2 {
		t.Errorf("Expected 2 results in history, but got %d", len(history))
	}

	// 验证历史记录内容
	if history[0].Name != "history_task1" {
		t.Errorf("Expected first result name 'history_task1', but got '%s'", history[0].Name)
	}

	if history[1].Name != "history_task2" {
		t.Errorf("Expected second result name 'history_task2', but got '%s'", history[1].Name)
	}
}

func TestDailyTaskQueue_move(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 保存原始时间
	now := time.Now()
	expectedTick := now.Hour()*60 + now.Minute()
	expectedDate := now.Format("2006-01-02")

	// 调用 move 方法
	tq.move()

	// 验证 tick 和 date 是否正确设置
	if tq.tick != expectedTick {
		t.Errorf("Expected tick %d, but got %d", expectedTick, tq.tick)
	}

	if tq.date != expectedDate {
		t.Errorf("Expected date '%s', but got '%s'", expectedDate, tq.date)
	}
}

func TestDailyTaskQueue_appendResult(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 清空历史记录
	tq.results = nil

	// 添加结果直到超过限制
	limit := 1000
	for i := 0; i < limit+10; i++ {
		result := dailyTaskResult{
			Name:    "test_task",
			Message: "test message",
		}
		tq.appendResult(result)
	}

	// 验证结果数量是否在限制范围内
	if len(tq.results) > limit {
		t.Errorf("Expected at most %d results, but got %d", limit, len(tq.results))
	}

	// 验证是否移除了旧的结果
	if len(tq.results) < limit-5 {
		t.Errorf("Expected at least %d results, but got %d", limit-5, len(tq.results))
	}
}

func TestDailyTaskQueue_caller(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 创建测试任务
	executed := false
	task := &dailyTask{
		Name: "caller_test",
		Fn: func() string {
			executed = true
			return "task executed"
		},
	}

	// 调用任务
	tq.caller(task, true, "")

	// 验证任务被执行
	if !executed {
		t.Error("Expected task to be executed")
	}

	// 验证任务状态
	if task.Running {
		t.Error("Expected task to not be running after execution")
	}

	// 验证结果被添加到历史记录
	if len(tq.results) != 1 {
		t.Errorf("Expected 1 result in history, but got %d", len(tq.results))
	}

	// 测试任务已经在运行的情况
	task.Running = true
	tq.results = nil // 清空之前的结果
	tq.caller(task, true, "")

	// 验证错误结果被添加到历史记录
	if len(tq.results) != 1 {
		t.Errorf("Expected 1 result in history, but got %d", len(tq.results))
	}

	if tq.results[0].Error == nil {
		t.Error("Expected error in result for running task")
	}
}

func TestNewDaily(t *testing.T) {
	// 创建新的 dailyTaskQueue
	d := NewDaily()

	// 验证初始化
	if d.started {
		t.Error("Expected daily task queue to not be started initially")
	}

	// 验证时间和日期被正确设置
	now := time.Now()
	expectedTick := now.Hour()*60 + now.Minute()
	expectedDate := now.Format("2006-01-02")

	// 允许一些时间差
	if d.tick < expectedTick-1 || d.tick > expectedTick+1 {
		t.Errorf("Expected tick around %d, but got %d", expectedTick, d.tick)
	}

	if d.date != expectedDate {
		t.Errorf("Expected date '%s', but got '%s'", expectedDate, d.date)
	}
}

func TestDailyTaskQueue_Start(t *testing.T) {
	// 创建新的 dailyTaskQueue 实例
	tq := &dailyTaskQueue{}

	// 验证初始状态
	if tq.started {
		t.Error("Expected daily task queue to not be started initially")
	}

	// 调用 Start 方法
	tq.Start()

	// 验证状态已更新
	if !tq.started {
		t.Error("Expected daily task queue to be marked as started")
	}

	// 再次调用 Start 方法，应该不会重复启动
	// 由于日志输出难以测试，我们主要验证不会出现异常
	tq.Start()

	// 等待一小段时间，让 goroutine 启动
	time.Sleep(10 * time.Millisecond)
}

func TestCancel(t *testing.T) {
	// 重置测试状态
	daily = dailyTaskQueue{}

	// 注册测试任务
	daily.RegisteTask("cancel_task", func() string { return "test" }, 540)

	// 取消任务
	canceled := Cancel("cancel_task")

	// 验证返回结果
	if len(canceled) != 1 {
		t.Errorf("Expected 1 canceled task, but got %d", len(canceled))
	}

	if canceled[0].Name != "cancel_task" {
		t.Errorf("Expected canceled task name 'cancel_task', but got '%s'", canceled[0].Name)
	}
}
