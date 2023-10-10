package unix

import "time"

func TimeNow() int64 {
	return time.Now().UnixNano()
}
