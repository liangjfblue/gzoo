package monitor

import (
	"fmt"
	"testing"
	"time"
)

func TestMachineInfo_ToMap(t *testing.T) {
	m := newMachine()
	for {
		select {
		case <-time.After(time.Second * 3):
			res := m.GetInfo()
			fmt.Println(res)
		}
	}
}
