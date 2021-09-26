package receiver

import "time"

type transferInfo struct {
	ReceivedFileBytesPackets uint64
	ApproximateNumOfPackets  uint64
	StartTime                time.Time
}
