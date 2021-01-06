package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Reporter interface {
	Report(Options) error
}

type respond struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (r *respond) String() string {
	return fmt.Sprintf("code:%d, msg:%s", r.Code, r.Msg)
}

type httpReporter struct {
	machineInfo *MachineInfo
}

var (
	errInfoEmpty = errors.New("report info empty")
)

func NewHttpReporter() Reporter {
	r := &httpReporter{}
	r.machineInfo = newMachine()
	return r
}

func (h *httpReporter) Report(opts Options) error {
	info := h.machineInfo.GetInfo()
	if info == nil {
		return errInfoEmpty
	}
	info["serviceName"] = opts.ServiceName
	info["portList"] = opts.ServicePort

	infoStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if opts.IsDebug {
		fmt.Println(string(infoStr))
	}

	//resp, err := http.Post(opts.HttpReportUrl, "application/json", strings.NewReader(string(infoStr)))
	//if err != nil {
	//	return err
	//}
	//defer resp.Body.Close()
	//
	//ret, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//
	//var respond respond
	//if err = json.Unmarshal(ret, &respond); err != nil {
	//	return err
	//}
	//
	//if respond.Code != 1 {
	//	return fmt.Errorf("report data error respond:%s", respond.String())
	//}

	return nil
}

type agentReporter struct {
	machineInfo *MachineInfo
}

func NewAgentReporter(opts Options) Reporter {
	r := &agentReporter{}
	r.machineInfo = newMachine()
	return r
}

func (r *agentReporter) Report(opts Options) error {
	info := r.machineInfo.GetInfo()
	if info == nil {
		return errInfoEmpty
	}
	info["serviceName"] = opts.ServiceName
	info["portList"] = opts.ServicePort

	infoStr, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if opts.IsDebug {
		fmt.Println(string(infoStr))
	}

	//todo agent tcp report

	return nil
}
