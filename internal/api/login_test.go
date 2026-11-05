package api

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123789456"), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("密码加密失败: %v", err)
	}
	password := string(hashedPassword)
	t.Log("123789456: " + password)
}
