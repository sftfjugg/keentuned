package modules

import (
	"fmt"
	"keentune/daemon/common/utils/http"
	"sync"
)

type SafeChan struct {
	C    chan struct{}
	once sync.Once
}

// NewSafeChan ...
func NewSafeChan() *SafeChan {
	return &SafeChan{C: make(chan struct{}, 1)}
}

// SafeStop ...
func (sc *SafeChan) SafeStop() {
	sc.once.Do(func() {
		close(sc.C)
	})
}

func terminate(host string) error {
	url := fmt.Sprintf("%v/terminate", host)
	return http.ResponseSuccess("GET", url, nil)
}

