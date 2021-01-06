package monitor

import "time"

var (
	defaultOptions = Options{
		Interval:        time.Second * 3,
		IsDebug:         true,
		AgentReportIp:   "127.0.0.1",
		AgentReportPort: 9090,
		HttpReportUrl:   "http://127.0.0.1:8080/monitor/data",
		ReporterFunc:    NewHttpReporter(), //默认http
	}
)

type Options struct {
	ServiceName     string        //服务名
	ServicePort     []int         //服务端口
	ReporterFunc    Reporter      //上报方式
	Interval        time.Duration //上报间隔
	IsDebug         bool          //上报调试方式, 终端打印
	AgentReportIp   string        //agent上报方式的ip
	AgentReportPort int           //agent上报方式的port
	HttpReportUrl   string        //http上报方式url
}

func MachineName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}

func MachinePort(servicePort []int) Option {
	return func(o *Options) {
		o.ServicePort = servicePort
	}
}

func ReporterFunc(r Reporter) Option {
	return func(o *Options) {
		o.ReporterFunc = r
	}
}

func Interval(interval time.Duration) Option {
	return func(o *Options) {
		o.Interval = interval
	}
}

func IsDebug(isDebug bool) Option {
	return func(o *Options) {
		o.IsDebug = isDebug
	}
}

func AgentReportIp(agentReportIp string) Option {
	return func(o *Options) {
		o.AgentReportIp = agentReportIp
	}
}

func AgentReportPort(agentReportPort int) Option {
	return func(o *Options) {
		o.AgentReportPort = agentReportPort
	}
}

func HttpReportUrl(httpReportUrl string) Option {
	return func(o *Options) {
		o.HttpReportUrl = httpReportUrl
	}
}
