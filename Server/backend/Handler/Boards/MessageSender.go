package boards

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

func PushMessageToCentral(frpAddress string, timeout time.Duration, maxRetryAttempts int, message *Board.BoardMessage) (*Board.BoardMessage, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 1
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		conn, err := net.DialTimeout("tcp", frpAddress, timeout)
		if err != nil {
			lastErr = err
			continue
		}

		data, err := json.Marshal(message)
		if err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		_, err = conn.Write(data)
		if err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		conn.SetReadDeadline(time.Now().Add(timeout))
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		conn.Close()
		if err != nil {
			lastErr = err
			continue
		}

		var response Board.BoardMessage
		if err := json.Unmarshal(buffer[:n], &response); err != nil {
			lastErr = err
			continue
		}

		return &response, nil
	}

	return nil, fmt.Errorf("PushMessageToCentral failed after %d attempts: %v", maxRetryAttempts, lastErr)
}
