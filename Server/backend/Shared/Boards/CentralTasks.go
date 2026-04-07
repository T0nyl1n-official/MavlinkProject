package boards

// CentralProgressChain 表示将要发送给树莓派 (CentralBoard) 的任务链结构
type CentralProgressChain struct {
	ChainID     string        `json:"chain_id"`
	Tasks       []CentralTask `json:"tasks"`
	CurrentTask int           `json:"current_task"`
	Status      string        `json:"status"`
}

// CentralTask 表示其中一个具体的任务
type CentralTask struct {
	TaskID  string                 `json:"task_id"`
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
	Status  string                 `json:"status"`
}
