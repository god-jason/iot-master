package hikvideo

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/url"
)

// 常量（完全对齐 Java）
const (
	MaxEncryptBlock = 117
	MaxDecryptBlock = 128
)

// ============================
// 私钥解析（PKCS8）
// ============================
func parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}

	keyInterface, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	priv, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not RSA private key")
	}

	return priv, nil
}

// ============================
// 私钥加密（模拟 Java Cipher ENCRYPT_MODE）
// ============================
func privateEncrypt(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	k := (priv.N.BitLen() + 7) / 8

	if len(data) > k-11 {
		return nil, fmt.Errorf("data too long")
	}

	// PKCS1Padding（和 Java 一致）
	em := make([]byte, k)
	em[0] = 0
	em[1] = 1

	psLen := k - len(data) - 3
	for i := 0; i < psLen; i++ {
		em[2+i] = 0xff
	}

	em[2+psLen] = 0
	copy(em[3+psLen:], data)

	// m^d mod n
	m := new(big.Int).SetBytes(em)
	c := new(big.Int).Exp(m, priv.D, priv.N)

	out := c.Bytes()
	if len(out) < k {
		tmp := make([]byte, k)
		copy(tmp[k-len(out):], out)
		out = tmp
	}

	return out, nil
}

// ============================
// 私钥解密（模拟 Java Cipher DECRYPT_MODE）
// ============================
func privateDecrypt(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	k := (priv.N.BitLen() + 7) / 8

	if len(data) != k {
		return nil, fmt.Errorf("invalid block size")
	}

	// c^d mod n
	c := new(big.Int).SetBytes(data)
	m := new(big.Int).Exp(c, priv.D, priv.N)

	out := m.Bytes()
	if len(out) < k {
		tmp := make([]byte, k)
		copy(tmp[k-len(out):], out)
		out = tmp
	}

	// 去掉 PKCS1Padding
	if out[0] != 0 || out[1] != 1 {
		return nil, fmt.Errorf("invalid padding")
	}

	i := 2
	for ; i < len(out); i++ {
		if out[i] == 0 {
			break
		}
	}

	return out[i+1:], nil
}

// ============================
// 加密接口（完全对齐 Java encryptByPrivateKey）
// ============================
func EncryptByPrivateKey(data string, secret string) (string, error) {

	// 1️⃣ URL Encode
	encoded := url.QueryEscape(data)

	// 2️⃣ 私钥
	priv, err := parsePrivateKey(secret)
	if err != nil {
		return "", err
	}
	
	dataBytes := []byte(encoded)

	var buffer bytes.Buffer

	// 3️⃣ 分段加密
	for i := 0; i < len(dataBytes); i += MaxEncryptBlock {
		end := i + MaxEncryptBlock
		if end > len(dataBytes) {
			end = len(dataBytes)
		}

		block := dataBytes[i:end]

		enc, err := privateEncrypt(priv, block)
		if err != nil {
			return "", err
		}

		buffer.Write(enc)
	}

	// 4️⃣ Base64
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// ============================
// 解密接口（完全对齐 Java decryptByPrivateKey）
// ============================
func DecryptByPrivateKey(data string, secret string) (string, error) {

	priv, err := parsePrivateKey(secret)
	if err != nil {
		return "", err
	}

	// 1️⃣ Base64 解码
	dataBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	// 2️⃣ 分段解密
	for i := 0; i < len(dataBytes); i += MaxDecryptBlock {
		end := i + MaxDecryptBlock
		if end > len(dataBytes) {
			end = len(dataBytes)
		}

		block := dataBytes[i:end]

		dec, err := privateDecrypt(priv, block)
		if err != nil {
			return "", err
		}

		buffer.Write(dec)
	}

	// 3️⃣ URL Decode
	result, err := url.QueryUnescape(buffer.String())
	if err != nil {
		return "", err
	}

	return result, nil
}
