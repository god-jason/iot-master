package mqtt

import (
	"strings"
	"testing"
)

func TestBuildTopic(t *testing.T) {
	cases := []struct{ id, suffix, want string }{
		{"dev1", PropertyReport, "$thing/dev1/property/report"},
		{"dev1", PropertySetResp, "$thing/dev1/property/set/response"},
		{"dev1", DatabaseLoad, "$thing/dev1/database/load"},
	}
	for _, c := range cases {
		if got := BuildTopic(c.id, c.suffix); got != c.want {
			t.Errorf("BuildTopic(%q, %q) = %q, want %q", c.id, c.suffix, got, c.want)
		}
	}
}

func TestDeviceIDFromTopic(t *testing.T) {
	cases := []struct {
		topic string
		want  string
		ok    bool
	}{
		{"$thing/dev1/property/report", "dev1", true},
		{"$thing/dev1/property/set/response", "dev1", true},
		{"dev1/property/report", "", false},
		{"$thing//property/report", "", false},
		{"$thing/dev1", "", false},
		{"$other/dev1/property/report", "", false},
	}
	for _, c := range cases {
		got, ok := deviceIDFromTopic(c.topic)
		if got != c.want || ok != c.ok {
			t.Errorf("deviceIDFromTopic(%q) = (%q, %v), want (%q, %v)", c.topic, got, ok, c.want, c.ok)
		}
	}
}

func TestMsgTime(t *testing.T) {
	if got := msgTime(0); got.IsZero() {
		t.Fatal("msgTime(0) 应返回当前时间")
	}
	ts := int64(1700000000000)
	if got := msgTime(ts); got.UnixMilli() != ts {
		t.Errorf("msgTime(%d).UnixMilli() = %d", ts, got.UnixMilli())
	}
}

func TestResponseManager(t *testing.T) {
	rm := NewResponseManager()

	ch, err := rm.Add("req-1")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := rm.Add("req-1"); err != ErrDuplicateRequest {
		t.Errorf("重复 Add 应返回 ErrDuplicateRequest, got %v", err)
	}

	if ok := rm.HandleResponse("missing", &ResponseMessage{ID: "missing"}); ok {
		t.Error("未登记的 id 不应匹配成功")
	}

	if ok := rm.HandleResponse("req-1", &ResponseMessage{ID: "req-1", Ret: 1}); !ok {
		t.Fatal("已登记的 id 应匹配成功")
	}
	select {
	case resp := <-ch:
		if resp == nil || resp.ID != "req-1" || resp.Ret != 1 {
			t.Errorf("收到的应答不正确: %+v", resp)
		}
	default:
		t.Fatal("应答未投递到通道")
	}

	// HandleResponse 匹配后应删除待匹配项，可再次 Add
	if _, err := rm.Add("req-1"); err != nil {
		t.Errorf("匹配后应可重新 Add, got %v", err)
	}
	rm.Remove("req-1")
}

func TestSubscriptionFilters(t *testing.T) {
	seen := make(map[string]bool)
	for _, sub := range subscriptions {
		if seen[sub.filter] {
			t.Errorf("重复订阅 filter: %s", sub.filter)
		}
		seen[sub.filter] = true

		if sub.handler == nil {
			t.Errorf("filter %s 缺少 handler", sub.filter)
		}
		if !strings.HasPrefix(sub.filter, TopicPrefix+"/+/") {
			t.Errorf("filter %s 不符合 $thing/+/x 形式", sub.filter)
		}
	}
}
