package lib

import (
	"math/rand"
	"time"
)

var chars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandomString(lenNum int) string {
	buf := make([]byte, lenNum)
	length := len(chars)
	rand.Seed(time.Now().UnixNano()) //重新播种，否则值不会变
	for i := 0; i < lenNum; i++ {
		buf[i] = chars[rand.Intn(length)] // +号拼接字符性能低，需要copy,频繁新增释放操作
	}
	return string(buf)

}
