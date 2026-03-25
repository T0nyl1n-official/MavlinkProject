package boards

type BoardType string

const (
	Type_Drone    BoardType = "Drone"
	Type_Control  BoardType = "Control"
	Type_LandNode BoardType = "CentralBoard"
)

// Board 结构体用于表示装载在飞控上的可编程板, 飞控板本身,
type Board struct {
	BoardID   string    `json:"board_id"`
	BoardType BoardType `json:"board_type"`
	BoardName string    `json:"board_name"`
	BoardDesc string    `json:"board_desc"`

	BoardStatus string `json:"board_status"`
	BoardConfig string `json:"board_config"`
	BoardIP     string `json:"board_ip"`
	BoardPort   string `json:"board_port"`

	IsOnline    bool `json:"is_online"`
	IsConnected bool `json:"is_connected"`
}
