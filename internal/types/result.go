package types

import "time"

// ResolveCallResult 统一转换设备控制调用结果为（是否成功, 描述），规则引擎与场景引擎共用：
// 调用错误优先；nil 响应视为异步下发成功；Ret==1 为设备执行成功。
func ResolveCallResult(resp *CallResult, err error) (bool, string) {
	if err != nil {
		return false, err.Error()
	}
	if resp == nil {
		return true, "指令已下发"
	}
	if resp.Ret == 1 {
		return true, "设备执行成功"
	}
	return false, resp.Message
}

// WaitUntilNextMinute 阻塞到下一个整分钟，使 ticker 触发点与 cron 的分钟边界对齐。
func WaitUntilNextMinute() {
	now := time.Now()
	next := now.Truncate(time.Minute).Add(time.Minute)
	time.Sleep(next.Sub(now))
}
