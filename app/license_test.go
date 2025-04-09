package app

import (
	"os"
	"testing"
)

func TestGenerateLic(t *testing.T) {
	lic := &License{
		AppId:     "test2",
		Owner:     "jason",
		Issuer:    "benyi",
		Issued:    "2025-10-20",
		Expire:    "2025-10-20",
		Cpuid:     "",
		Mac:       "",
		Hosts:     nil,
		Signature: "",
	}

	lic.Sign(priKey)
	data, err := lic.Encode()
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile("test2.lic", []byte(data), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}
