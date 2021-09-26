package sender

import "time"

type transferInfo struct {
	SentFileBytesPackets    uint64
	ApproximateNumOfPackets uint64
	StartTime               time.Time
}
