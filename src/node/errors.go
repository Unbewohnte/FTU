package node

import "fmt"

var (
	ErrorNotConnected error = fmt.Errorf("not connected")
	ErrorSentAll      error = fmt.Errorf("sent the whole file")
)
