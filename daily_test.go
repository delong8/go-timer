package timer

import (
	"testing"
	"time"
)

func TestNewDaily(t *testing.T) {
	tq := NewDaily()

	// 检查是否正确初始化
	if tq.started {
		t.Error("Expected started to be false, but got true")
	}

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

func TestDailyTaskQueue_CancelTask(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 添加测试任务
	err := tq.RegisteTask("task1", func() string { return "task1" }, 0)
	if err != nil {
		t.Fatalf("Failed to register task1: %v", err)
	}

	err = tq.RegisteTask("task2", func() string { return "task2" }, 100)
	if err != nil {
		t.Fatalf("Failed to register task2: %v", err)
	}

	err = tq.RegisteTask("task3", func() string { return "task3" }, 200)
	if err != nil {
		t.Fatalf("Failed to register task3: %v", err)
	}

	// 取消 task1，应该取消除了 task1 之外的所有任务
	canceled := tq.CancelTask("task1")
	if len(canceled) != 2 {
		t.Errorf("Expected 2 canceled tasks, but got %d", len(canceled))
	}

	// 检查返回的取消任务列表
	expectedCanceled := map[string]bool{
		"task2": true,
		"task3": true,
	}

	for _, c := range canceled {
		if !expectedCanceled[c.Name] {
			t.Errorf("Unexpected canceled task: %s", c.Name)
		}
		// 验证这些任务在队列中也被标记为已取消
		for _, task := range tq.tasks {
			if task.Name == c.Name {
				if !task.Canceled {
					t.Errorf("Expected task %s to be canceled", task.Name)
				}
				break
			}
		}
	}

	// 验证未被取消的任务
	for _, task := range tq.tasks {
		if task.Name == "task1" {
			if task.Canceled {
				t.Errorf("Expected task %s not to be canceled", task.Name)
			}
			break
		}
	}
}

func TestDailyTaskQueue_ShouldRun(t *testing.T) {
	tq := &dailyTaskQueue{}
	tq.move()

	// 测试未到执行时间的任务
	futureTask := dailyTask{
		Name:      "future_task",
		RunAtTick: tq.tick + 10,
		RunAtDate: "",
	}

	if tq.shouldRun(futureTask) {
		t.Error("Expected future task not to run, but it should run")
	}

	// 测试已执行的任务
	executedTask := dailyTask{
		Name:      "executed_task",
		RunAtTick: tq.tick - 10,
		RunAtDate: tq.date,
	}

	if tq.shouldRun(executedTask) {
		t.Error("Expected executed task not to run again, but it should run")
	}

	// 测试已取消的任务
	canceledTask := dailyTask{
		Name:      "canceled_task",
		RunAtTick: tq.tick - 5,
		RunAtDate: "",
		Canceled:  true,
	}

	if tq.shouldRun(canceledTask) {
		t.Error("Expected canceled task not to run, but it should run")
	}

	// 测试应该执行的任务
	readyTask := dailyTask{
		Name:      "ready_task",
		RunAtTick: tq.tick - 5,
		RunAtDate: "",
	}

	if !tq.shouldRun(readyTask) {
		t.Error("Expected ready task to run, but it should not run")
	}
}

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
	task := &dailyTask{
		Name:      "test_task",
		Fn:        func() string { return "success" },
		RunAtTick: 0,
		RunAtDate: "",
	}

	tq.caller(task, true, "") // 手动触发

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

	if !tq.results[0].Manually {
		t.Error("Expected result to be manually executed")
	}

	// 测试正在运行的任务
	task.Running = true
	tq.results = []dailyTaskResult{} // 重置结果
	tq.caller(task, true, "")        // 手动触发

	// 检查结果是否被记录为错误
	if len(tq.results) != 1 {
		t.Errorf("Expected 1 result, but got %d", len(tq.results))
	}

	if tq.results[0].Error == nil {
		t.Error("Expected error for running task, but got none")
	}

	// 测试任务执行日期更新
	task.Running = false
	task.RunAtDate = ""
	tq.results = []dailyTaskResult{} // 重置结果
	testDate := "2023-01-01"
	tq.caller(task, false, testDate) // 自动触发

	if task.RunAtDate != testDate {
		t.Errorf("Expected RunAtDate to be %s, but got %s", testDate, task.RunAtDate)
	}
}
// TestDailyTaskQueue_Caller_Panic 测试 caller 函数中发生 panic 的情况
func TestDailyTaskQueue_Caller_Panic(t *testing.T) {
	tq := &dailyTaskQueue{}
	
	// 创建一个会 panic 的任务函数
	task := &dailyTask{
		Name:      "panic_task",
		Fn:        func() string { panic("test panic") },
		RunAtTick: 0,
		RunAtDate: "",
	}
	
	// 调用会 panic 的任务
	tq.caller(task, true, "")
	
	// 检查结果是否被记录
	if len(tq.results) != 1 {
		t.Errorf("Expected 1 result, but got %d", len(tq.results))
	}
	
	if tq.results[0].Name != "panic_task" {
		t.Errorf("Expected result name to be 'panic_task', but got %s", tq.results[0].Name)
	}
	
	// 检查任务状态是否被重置
	if task.Running {
		t.Error("Expected task Running to be false after panic")
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
	tq.caller(task, true, "") // 手动触发

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

func TestDailyTaskQueue_Looper(t *testing.T) {
	// 这个测试比较复杂，因为 looper 是一个无限循环
	// 我们可以通过检查任务是否被正确执行来间接测试它

	tq := &dailyTaskQueue{}
	tq.move()

	// 创建一个 channel 来接收任务执行结果
	resultChan := make(chan string, 2)

	// 注册一个应该立即执行的任务（过去的tick）
	err := tq.RegisteTask("immediate_task", func() string {
		resultChan <- "immediate executed"
		return "immediate success"
	}, tq.tick-1)

	if err != nil {
		t.Fatalf("Failed to register immediate task: %v", err)
	}

	// 注册一个将来执行的任务
	err = tq.RegisteTask("future_task", func() string {
		resultChan <- "future executed"
		return "future success"
	}, tq.tick+1)

	if err != nil {
		t.Fatalf("Failed to register future task: %v", err)
	}

	// 模拟一次循环迭代
	tq.move()
	var tasks []*dailyTask
	for _, t := range tq.tasks {
		if t.Canceled {
			continue
		}
		tasks = append(tasks, t)
	}
	tq.tasks = tasks
	// 重置所有任务的RunAtDate，确保它们可以被执行
	for _, task := range tq.tasks {
		task.RunAtDate = ""
	}

	for _, task := range tq.tasks {
		if tq.shouldRun(*task) {
			tq.caller(task, false, tq.date)
		}
	}

	// 检查立即执行的任务是否被执行
	select {
	case result := <-resultChan:
		if result != "immediate executed" {
			t.Errorf("Expected 'immediate executed', but got %s", result)
		}
	case <-time.After(1 * time.Second):
		t.Error("Immediate task execution timed out")
	}

	// 检查历史记录
	history := tq.History()
	if len(history) < 1 {
		t.Errorf("Expected at least 1 history item, but got %d", len(history))
	}
}
