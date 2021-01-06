package monitor

import (
	"errors"
	"log"
	"time"
)

type Monitor interface {
	Init(...Option) Monitor
	Run() error
	Stop()
}

type Option func(*Options)

var _ Monitor = &monitor{}

var (
	DefaultMonitor = NewMonitor()
)

type monitor struct {
	report Reporter
	opts   Options
	stop   chan struct{}
}

func NewMonitor() Monitor {
	return &monitor{}
}

func (m *monitor) Init(opts ...Option) Monitor {
	m.opts = defaultOptions
	for _, o := range opts {
		o(&m.opts)
	}

	m.report = m.opts.ReporterFunc
	m.stop = make(chan struct{}, 1)

	log.Println("monitor init")

	return m
}

func (m *monitor) Run() error {
	var err error

	if m.report == nil {
		return errors.New("report function nil")
	}

	t := time.NewTicker(m.opts.Interval)
	go func() {
		for {
			select {
			case <-t.C:
				if err = m.report.Report(m.opts); err != nil {
					log.Println("report data err:", err.Error())
				}
			case <-m.stop:
				log.Println("monitor stop")
				return
			}
		}
	}()

	return nil
}

func (m *monitor) Stop() {
	m.stop <- struct{}{}
}
