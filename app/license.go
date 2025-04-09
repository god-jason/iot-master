package app

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type License struct {
	AppId     string   `json:"app_id,omitempty"` //应用ID
	Owner     string   `json:"owner,omitempty"`  //拥有者
	Issuer    string   `json:"issuer,omitempty"` //发行者
	Issued    string   `json:"issued,omitempty"` //发布日期
	Expire    string   `json:"expire,omitempty"` //失效日期
	Cpuid     string   `json:"cpuid,omitempty"`  //CPUID
	Mac       string   `json:"mac,omitempty"`    //网卡ID
	Hosts     []string `json:"hosts,omitempty"`  //域名
	Signature string   `json:"sign,omitempty"`   //签名
}

func (l *License) String() string {
	hosts := strings.Join(l.Hosts, ",")
	ss := []string{l.AppId, l.Owner, l.Issuer, l.Issued, l.Expire, l.Cpuid, l.Mac, hosts}
	return strings.Join(ss, ",")
}

func (l *License) Encode() (string, error) {
	buf, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(buf)
	return b64, nil
}

func (l *License) Decode(lic string) error {
	buf, err := base64.StdEncoding.DecodeString(lic)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, l)
}

func (l *License) Sign(privateKey []byte) {
	data := l.String()
	sign := ed25519.Sign(privateKey, []byte(data))
	l.Signature = hex.EncodeToString(sign)
}

func (l *License) Verify(publicKey []byte) error {
	sign, err := hex.DecodeString(l.Signature)
	if err != nil {
		return err
	}

	//检查签名
	data := l.String()
	ret := ed25519.Verify(publicKey, []byte(data), sign)
	if ret == false {
		return errors.New("license verify error")
	}

	//检查失效期
	date, err := time.Parse(time.DateOnly, l.Expire)
	if err != nil {
		return err
	}
	if date.Before(time.Now()) {
		return errors.New("license expired")
	}

	return nil
}
