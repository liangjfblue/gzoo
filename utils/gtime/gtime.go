/*
@Time : 2021/1/24 23:46
@Author : liangjiefan
*/
package gtime

import "time"

// CostTime cal cost time(ms)
func CostTime(f func()) int {
	// Nanosecond 单调递增, 保证计算正确
	startTime := time.Now().Nanosecond()
	f()
	endTime := time.Now().Nanosecond()
	return (endTime - startTime) / 1e6
}
