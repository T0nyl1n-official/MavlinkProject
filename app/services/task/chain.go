package task

import (
	"fmt"
	"time"

	"MavlinkProject_Board/app/services/mavlink"
)

type ProgressChain struct {
	ChainID       string    `json:"chain_id"`
	Tasks         []Task    `json:"tasks"`
	CurrentTask   int       `json:"current_task"`
	Status        string    `json:"status"`
	StartTime     int64     `json:"start_time"`
	EndTime       int64     `json:"end_time"`
	AssignedDrone string    `json:"assigned_drone"`
}

type Task struct {
	TaskID     string                 `json:"task_id"`
	Command    string                 `json:"command"`
	Data       map[string]interface{} `json:"data"`
	Status     string                 `json:"status"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Timeout    int                    `json:"timeout"`
	StartTime  int64                  `json:"start_time"`
	EndTime    int64                  `json:"end_time"`
}

var taskChains = make(map[string]*ProgressChain)

func CreateTaskChain(tasks []struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
	Timeout int                    `json:"timeout"`
}) (*ProgressChain, error) {
	chainID := fmt.Sprintf("chain_%d", time.Now().UnixNano())

	chain := &ProgressChain{
		ChainID:     chainID,
		Tasks:       make([]Task, len(tasks)),
		CurrentTask: 0,
		Status:      "pending",
		StartTime:   time.Now().Unix(),
	}

	for i, t := range tasks {
		chain.Tasks[i] = Task{
			TaskID:     fmt.Sprintf("task_%s_%d", chainID, i),
			Command:    t.Command,
			Data:       t.Data,
			Status:     "pending",
			MaxRetries: 3,
			Timeout:    t.Timeout,
		}
	}

	taskChains[chainID] = chain

	// 启动任务链执行
	go executeTaskChain(chain)

	return chain, nil
}

func ListTaskChains() []*ProgressChain {
	chains := make([]*ProgressChain, 0, len(taskChains))
	for _, chain := range taskChains {
		chains = append(chains, chain)
	}
	return chains
}

func GetTaskChain(chainID string) (*ProgressChain, error) {
	chain, exists := taskChains[chainID]
	if !exists {
		return nil, fmt.Errorf("task chain not found")
	}
	return chain, nil
}

func executeTaskChain(chain *ProgressChain) {
	chain.Status = "running"

	for i, task := range chain.Tasks {
		chain.CurrentTask = i
		task.Status = "running"
		task.StartTime = time.Now().Unix()

		err := executeTask(&task)
		if err != nil {
			task.Status = "failed"
			task.EndTime = time.Now().Unix()
			chain.Status = "failed"
			chain.EndTime = time.Now().Unix()
			return
		}

		task.Status = "completed"
		task.EndTime = time.Now().Unix()
	}

	chain.Status = "completed"
	chain.EndTime = time.Now().Unix()
}

func executeTask(task *Task) error {
	switch task.Command {
	case "takeoff":
		return mavlink.TakeOff(task.Data)
	case "land":
		return mavlink.Land(task.Data)
	case "goto", "goto_location":
		return mavlink.GoTo(task.Data)
	case "return_to_home", "rtl":
		return mavlink.ReturnToHome()
	case "survey":
		return mavlink.Survey(task.Data)
	case "survey_grid":
		return mavlink.SurveyGrid(task.Data)
	case "orbit":
		return mavlink.Orbit(task.Data)
	case "take_photo":
		return mavlink.TakePhoto()
	case "start_video":
		return mavlink.StartVideo()
	case "stop_video":
		return mavlink.StopVideo()
	case "set_mode":
		return mavlink.SetMode(task.Data)
	default:
		return fmt.Errorf("unknown command: %s", task.Command)
	}
}
