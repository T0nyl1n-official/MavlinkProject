package unit

import (
	"testing"

	"MavlinkProject_Board/app/services/task"
)

func TestCreateTaskChain(t *testing.T) {
	// 构造测试任务
	testTasks := []struct {
		Command string                 `json:"command"`
		Data    map[string]interface{} `json:"data"`
		Timeout int                    `json:"timeout"`
	}{
		{
			Command: "takeoff",
			Data: map[string]interface{}{
				"altitude": 10.0,
			},
			Timeout: 30,
		},
		{
			Command: "goto",
			Data: map[string]interface{}{
				"latitude":  22.543123,
				"longitude": 114.052345,
				"altitude":  15.0,
			},
			Timeout: 60,
		},
		{
			Command: "land",
			Data: map[string]interface{}{
				"latitude":  22.543123,
				"longitude": 114.052345,
				"altitude":  0.0,
			},
			Timeout: 30,
		},
	}

	// 创建任务链
	chain, err := task.CreateTaskChain(testTasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 验证任务链
	if chain.ChainID == "" {
		t.Error("Expected non-empty chain ID")
	}
	if len(chain.Tasks) != len(testTasks) {
		t.Errorf("Expected %d tasks, got %d", len(testTasks), len(chain.Tasks))
	}
	if chain.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", chain.Status)
	}

	// 验证任务
	for i, task := range chain.Tasks {
		if task.TaskID == "" {
			t.Errorf("Expected non-empty task ID for task %d", i)
		}
		if task.Command != testTasks[i].Command {
			t.Errorf("Expected command '%s' for task %d, got '%s'", testTasks[i].Command, i, task.Command)
		}
		if task.Status != "pending" {
			t.Errorf("Expected status 'pending' for task %d, got '%s'", i, task.Status)
		}
	}

	t.Logf("Task chain created successfully: %s", chain.ChainID)
}

func TestListTaskChains(t *testing.T) {
	// 先创建一个任务链
	testTasks := []struct {
		Command string                 `json:"command"`
		Data    map[string]interface{} `json:"data"`
		Timeout int                    `json:"timeout"`
	}{
		{
			Command: "takeoff",
			Data: map[string]interface{}{
				"altitude": 10.0,
			},
			Timeout: 30,
		},
	}

	_, err := task.CreateTaskChain(testTasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 列出任务链
	chains := task.ListTaskChains()
	if len(chains) == 0 {
		t.Error("Expected at least one task chain")
	}

	t.Logf("Found %d task chains", len(chains))
	for _, chain := range chains {
		t.Logf("Chain: %s, Status: %s, Tasks: %d", chain.ChainID, chain.Status, len(chain.Tasks))
	}
}

func TestGetTaskChain(t *testing.T) {
	// 创建一个任务链
	testTasks := []struct {
		Command string                 `json:"command"`
		Data    map[string]interface{} `json:"data"`
		Timeout int                    `json:"timeout"`
	}{
		{
			Command: "takeoff",
			Data: map[string]interface{}{
				"altitude": 10.0,
			},
			Timeout: 30,
		},
	}

	chain, err := task.CreateTaskChain(testTasks)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// 获取任务链
	retrievedChain, err := task.GetTaskChain(chain.ChainID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if retrievedChain.ChainID != chain.ChainID {
		t.Errorf("Expected chain ID '%s', got '%s'", chain.ChainID, retrievedChain.ChainID)
	}
	if len(retrievedChain.Tasks) != len(chain.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(chain.Tasks), len(retrievedChain.Tasks))
	}

	t.Logf("Task chain retrieved successfully: %s", retrievedChain.ChainID)
}
