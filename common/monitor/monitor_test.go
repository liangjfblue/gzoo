package monitor

import (
	"testing"
	"time"
)

func Test_defaultMonitor_Run(t *testing.T) {
	err := DefaultMonitor.Init(
		MachineName("test"),
		MachinePort([]int{9091}),
	).Run()
	defer DefaultMonitor.Stop()

	t.Log(err)

	select {
	case <-time.After(time.Second * 10):
	}
}

func Test_monitor_Run(t *testing.T) {
	err := NewMonitor().Init(
		MachineName("test"),
		MachinePort([]int{9091}),
		ReporterFunc(NewHttpReporter()),
		Interval(time.Second*3),
		IsDebug(true),
		HttpReportUrl("http://127.0.0.1:9090/monitor/data"),
	).Run()

	t.Log(err)

	select {
	case <-time.After(time.Second * 10):
	}
}
