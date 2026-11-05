package broker

import (
	"testing"
)

func TestParseTopic(t *testing.T) {
	cases := []struct {
		topic string
		dev   string
		suf   string
		ok    bool
	}{
		{"$thing/dev1/property/report", "dev1", "property/report", true},
		{"$thing/dev1/property/set/response", "dev1", "property/set/response", true},
		{"$thing//property/report", "", "", false},
		{"$thing/dev1", "", "", false},
		{"dev1/property/report", "", "", false},
		{"$other/dev1/property/report", "", "", false},
	}
	for _, c := range cases {
		dev, suf, ok := parseTopic(c.topic)
		if dev != c.dev || suf != c.suf || ok != c.ok {
			t.Errorf("parseTopic(%q) = (%q, %q, %v), want (%q, %q, %v)",
				c.topic, dev, suf, ok, c.dev, c.suf, c.ok)
		}
	}
}

func TestMsgTime(t *testing.T) {
	if msgTime(0).IsZero() {
		t.Fatal("msgTime(0) 应返回当前时间")
	}
	const ts = int64(1700000000000)
	if got := msgTime(ts).UnixMilli(); got != ts {
		t.Errorf("msgTime(%d).UnixMilli() = %d", ts, got)
	}
}

func TestCheckRet(t *testing.T) {
	if err := checkRet(nil); err == nil {
		t.Error("nil 应答应返回错误")
	}
	if err := checkRet(&ResponseMessage{Ret: 1}); err != nil {
		t.Errorf("ret=1 应成功, got %v", err)
	}
	if err := checkRet(&ResponseMessage{Ret: 0, Message: "failed"}); err == nil {
		t.Error("ret=0 应返回错误")
	}
}

func TestHandleResponseMatch(t *testing.T) {
	s := NewService()

	ch, err := s.rm.Add("req-1")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	payload := []byte(`{"id":"req-1","ret":1,"message":"success","data":{"v":1}}`)
	s.handleResponse("dev1", PropertySetResp, payload)

	select {
	case resp := <-ch:
		if resp == nil || resp.ID != "req-1" || resp.Ret != 1 {
			t.Errorf("应答内容不正确: %+v", resp)
		}
	default:
		t.Fatal("应答未投递到等待中的请求")
	}
}

func TestHandleResponseInvalidPayload(t *testing.T) {
	s := NewService()
	// 非法 JSON 不应 panic，也不应匹配任何请求
	s.handleResponse("dev1", PropertyGetResp, []byte(`{`))
}
