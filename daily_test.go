package timer

import (
	"testing"
	"time"
)

func TestDailyTaskQueue_Move(t *testing.T) {
	tq := &dailyTaskQueue{}
	tq.move()

	// 检查 tick 是否在有效范围内
	if tq.tick < 0 || tq.tick > 1439 {
		t.Errorf("Expected tick to be between 0 and 1439, but got %d", tq.tick)
	}

	// 检查日期格式是否正确
	expectedDate := time.Now().Format("2006-01-02")
	if tq.date != expectedDate {
		t.Errorf("Expected date format to be YYYY-MM-DD, got %s", tq.date)
	}
}

func TestDailyTaskQueue_Caller(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 测试正常执行的任务
	task := dailyTask{
		Name:      "test_task",
		Fn:        func() string { return "success" },
		RunAtTick: 0,
		RunAtDate: "",
	}

	tq.caller(task, true) // 手动触发

	// 检查结果是否被记录
	if len(tq.results) != 1 {
		t.Errorf("Expected 1 result, but got %d", len(tq.results))
	}

	if tq.results[0].Name != "test_task" {
		t.Errorf("Expected result name to be 'test_task', but got %s", tq.results[0].Name)
	}

	if tq.results[0].Message != "success" {
		t.Errorf("Expected result message to be 'success', but got %s", tq.results[0].Message)
	}

	// 测试正在运行的任务
	task.Running = true
	tq.results = []dailyTaskResult{} // 重置结果
	tq.caller(task, true)            // 手动触发

	// 检查结果是否被记录为错误
	if len(tq.results) != 1 {
		t.Errorf("Expected 1 result, but got %d", len(tq.results))
	}

	if tq.results[0].Error == nil {
		t.Error("Expected error for running task, but got none")
	}
}

func TestDailyTaskQueue_RunTask(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 添加一个测试任务
	err := tq.RegisteTask("test_task", func() string { return "test" }, 0)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	// 测试正常执行任务
	err = tq.RunTask("test_task")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// 测试执行不存在的任务
	err = tq.RunTask("nonexistent_task")
	if err == nil {
		t.Error("Expected error for nonexistent task, but got none")
	}
}

func TestDailyTaskQueue_Status(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 添加测试任务
	err := tq.RegisteTask("test_task", func() string { return "test" }, 0)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	// 检查状态
	tasks := tq.Status()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, but got %d", len(tasks))
	}

	if tasks[0].Name != "test_task" {
		t.Errorf("Expected task name to be 'test_task', but got %s", tasks[0].Name)
	}
}

func TestDailyTaskQueue_History(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 添加并执行一个测试任务
	err := tq.RegisteTask("test_task", func() string { return "test" }, 0)
	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	task := tq.tasks[0]
	tq.caller(task, true) // 手动触发

	// 检查历史记录
	history := tq.History()
	if len(history) != 1 {
		t.Errorf("Expected 1 history item, but got %d", len(history))
	}

	if history[0].Name != "test_task" {
		t.Errorf("Expected history name to be 'test_task', but got %s", history[0].Name)
	}
}
func TestDailyTaskQueue_Start(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 第一次调用 Start
	tq.Start()

	// 检查 started 标志是否设置为 true
	if !tq.started {
		t.Error("Expected started flag to be true after calling Start()")
	}

	// 再次调用 Start，应该不会重复启动
	tq.Start()

	// 检查 started 标志仍然为 true
	if !tq.started {
		t.Error("Expected started flag to remain true after calling Start() again")
	}
}

// TestDailyTaskExecution 测试任务执行的完整流程
func TestDailyTaskExecution(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 创建一个 channel 来接收任务执行结果
	resultChan := make(chan string, 1)

	// 注册一个测试任务
	err := tq.RegisteTask("execution_test", func() string {
		resultChan <- "task executed"
		return "success"
	}, 0)

	if err != nil {
		t.Fatalf("Failed to register task: %v", err)
	}

	// 手动执行任务
	err = tq.RunTask("execution_test")
	if err != nil {
		t.Fatalf("Failed to run task: %v", err)
	}

	// 等待任务执行结果
	select {
	case result := <-resultChan:
		if result != "task executed" {
			t.Errorf("Expected 'task executed', but got %s", result)
		}
	case <-time.After(1 * time.Second):
		t.Error("Task execution timed out")
	}

	// 检查历史记录
	history := tq.History()
	if len(history) != 1 {
		t.Errorf("Expected 1 history item, but got %d", len(history))
	}

	if history[0].Name != "execution_test" {
		t.Errorf("Expected history name to be 'execution_test', but got %s", history[0].Name)
	}

	if history[0].Message != "success" {
		t.Errorf("Expected history message to be 'success', but got %s", history[0].Message)
	}

	if !history[0].Manually {
		t.Error("Expected history to show manually executed task")
	}
}

func TestDailyTaskQueue_appendResult(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 添加足够多的结果来测试循环缓冲区
	for i := 0; i < 1001; i++ {
		result := dailyTaskResult{
			Name:    "test_task",
			Message: "test",
		}
		tq.appendResult(result)
	}

	// 检查结果数量是否正确（应该限制在1000个）
	// 当超过1000个结果时，会移除前两个结果，所以应该是999个
	if len(tq.results) != 999 {
		t.Errorf("Expected 999 results, but got %d", len(tq.results))
	}
}
