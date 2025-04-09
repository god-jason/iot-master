package app

import "testing"

func TestPack(t *testing.T) {
	err := Pack("../apps", "test2.app")
	if err != nil {
		t.Error(err)
	}
	//_ = Check("app.zip", "license")
}
