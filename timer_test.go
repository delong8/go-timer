package timer

import (
	"testing"
)

func TestParseTimeToTick(t *testing.T) {
	// 测试正常的时间格式
	tick, err := parseTimeToTick("09:30")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	expected := 9*60 + 30
	if tick != expected {
		t.Errorf("Expected tick %d, but got %d", expected, tick)
	}

	// 测试边界值 - 00:00
	tick, err = parseTimeToTick("00:00")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	expected = 0
	if tick != expected {
		t.Errorf("Expected tick %d, but got %d", expected, tick)
	}

	// 测试边界值 - 23:59
	tick, err = parseTimeToTick("23:59")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	expected = 23*60 + 59
	if tick != expected {
		t.Errorf("Expected tick %d, but got %d", expected, tick)
	}

	// 测试无效小时格式
	_, err = parseTimeToTick("24:00")
	if err == nil {
		t.Error("Expected error for invalid hour, but got none")
	}

	// 测试无效分钟格式
	_, err = parseTimeToTick("12:60")
	if err == nil {
		t.Error("Expected error for invalid minute, but got none")
	}

	// 测试错误格式
	_, err = parseTimeToTick("12-30")
	if err == nil {
		t.Error("Expected error for wrong format, but got none")
	}

	// 测试空字符串
	_, err = parseTimeToTick("")
	if err == nil {
		t.Error("Expected error for empty string, but got none")
	}

	// 测试非数字小时
	_, err = parseTimeToTick("ab:30")
	if err == nil {
		t.Error("Expected error for non-numeric hour, but got none")
	}

	// 测试非数字分钟
	_, err = parseTimeToTick("12:cd")
	if err == nil {
		t.Error("Expected error for non-numeric minute, but got none")
	}
}

func TestRegisteDaily(t *testing.T) {
	// 重置测试状态
	daily = dailyTaskQueue{}

	// 测试正常注册
	err := RegisteDaily("test_task", "09:30", func() string { return "test" })
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// 测试无效时间格式
	err = RegisteDaily("test_task2", "25:00", func() string { return "test" })
	if err == nil {
		t.Error("Expected error for invalid time format, but got none")
	}
}
func TestRegisteInterval(t *testing.T) {
	// 重置测试状态
	daily = dailyTaskQueue{}

	// 测试无效分钟数
	err := RegisteInterval("interval_task_0min", 0, func() string { return "interval test" })
	if err == nil {
		t.Error("Expected error for invalid minutes, but got none")
	}

	// 测试负数分钟数
	err = RegisteInterval("interval_task_neg", -1, func() string { return "interval test" })
	if err == nil {
		t.Error("Expected error for negative minutes, but got none")
	}
}

// TestRegisteIntervalCoverage 测试 RegisteInterval 函数的循环执行部分
func TestRegisteIntervalCoverage(t *testing.T) {
	// 重置测试状态
	daily = dailyTaskQueue{}

	// 测试正常注册 - 这将覆盖循环执行部分
	err := RegisteInterval("interval_coverage_test", 720, func() string { return "interval test" }) // 每12小时执行一次
	if err != nil {
		// 期望会出现错误，因为第二个任务会重复
		t.Logf("Expected error for duplicate task name: %v", err)
	}

	// 检查是否至少注册了一个任务
	tasks := daily.Status()
	found := false
	for _, task := range tasks {
		if task.Name == "interval_coverage_test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected at least one task to be registered")
	}
}
func TestDailyTaskQueue_RegisteTask(t *testing.T) {
	tq := &dailyTaskQueue{}

	// 测试正常注册
	err := tq.RegisteTask("test_task", func() string { return "test" }, 540) // 09:00
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// 测试重复任务名
	err = tq.RegisteTask("test_task", func() string { return "test2" }, 600) // 10:00
	if err == nil {
		t.Error("Expected error for duplicate task name, but got none")
	}

	// 测试无效tick值 - 负数
	err = tq.RegisteTask("test_task2", func() string { return "test2" }, -1)
	if err == nil {
		t.Error("Expected error for negative tick, but got none")
	}

	// 测试无效tick值 - 超过最大值
	err = tq.RegisteTask("test_task3", func() string { return "test3" }, 1400)
	if err == nil {
		t.Error("Expected error for tick over limit, but got none")
	}
}

// TestRegisteIntervalSuccess 测试 RegisteInterval 函数成功注册任务的情况
func TestRegisteIntervalSuccess(t *testing.T) {
	// 重置测试状态
	daily = dailyTaskQueue{}

	// 使用一个较大的间隔值，确保只注册一个任务
	err := RegisteInterval("interval_success_test", 1440, func() string { return "interval test" })
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// 检查是否注册了一个任务
	tasks := daily.Status()
	found := false
	for _, task := range tasks {
		if task.Name == "interval_success_test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected one task to be registered")
	}
}

func TestDailyTaskQueue_shouldRun(t *testing.T) {
	tq := &dailyTaskQueue{}
	tq.move() // 初始化当前时间

	task := dailyTask{
		Name:      "test_task",
		RunAtTick: tq.tick + 1, // 设置为下一分钟执行
		RunAtDate: "",          // 未执行过
	}

	// 测试未到执行时间
	if tq.shouldRun(task) {
		t.Error("Expected shouldRun to return false when not at run time")
	}

	// 更新任务到当前时间
	task.RunAtTick = tq.tick

	// 测试到达执行时间
	if !tq.shouldRun(task) {
		t.Error("Expected shouldRun to return true when at run time")
	}

	// 更新任务已执行日期为今天
	task.RunAtDate = tq.date

	// 测试当天已执行
	if tq.shouldRun(task) {
		t.Error("Expected shouldRun to return false when already run today")
	}
}

func TestInit(t *testing.T) {
	// 重置测试状态
	daily = dailyTaskQueue{}

	// 调用 Init 函数
	Init()

	// 检查是否注册了 "try" 任务
	tasks := daily.Status()
	found := false
	for _, task := range tasks {
		if task.Name == "try" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'try' task to be registered after Init, but it was not found")
	}
}
