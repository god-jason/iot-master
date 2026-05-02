package hikvideo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"math/big"
	"net/url"
)

const (
	MaxEncryptBlock = 117
	MaxDecryptBlock = 128
)

// EncryptByPrivateKey 使用RSA私钥加密数据
func EncryptByPrivateKey(data, privateKeyPEM string) (string, error) {
	// 获取私钥
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}

	// URL编码数据
	encodedData := url.QueryEscape(data)

	// 分块加密
	var encryptedData []byte
	for i := 0; i < len(encodedData); i += MaxEncryptBlock {
		end := i + MaxEncryptBlock
		if end > len(encodedData) {
			end = len(encodedData)
		}

		chunk := []byte(encodedData[i:end])
		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, chunk)
		if err != nil {
			return "", err
		}

		encryptedData = append(encryptedData, encryptedChunk...)
	}

	// Base64编码
	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// 私钥加密（模拟 RSA_private_encrypt）
func RsaPrivateEncrypt(privateKeyStr string, data string) (string, error) {
	// 获取私钥
	privateKey, err := parsePrivateKey(privateKeyStr)
	if err != nil {
		return "", err
	}
	// URL encode（和你C++一致）
	encodedData := url.QueryEscape(data)
	dataBytes := []byte(encodedData)

	keySize := privateKey.Size()
	maxBlock := keySize - 11 // PKCS1 padding

	var encrypted []byte

	for i := 0; i < len(dataBytes); i += maxBlock {
		end := i + maxBlock
		if end > len(dataBytes) {
			end = len(dataBytes)
		}

		blockData := dataBytes[i:end]

		// === 核心：手动做“私钥加密” ===
		m := new(big.Int).SetBytes(blockData)

		// c = m^d mod n
		c := new(big.Int).Exp(m, privateKey.D, privateKey.N)

		// 补齐长度
		out := c.Bytes()
		if len(out) < keySize {
			padding := make([]byte, keySize-len(out))
			out = append(padding, out...)
		}

		encrypted = append(encrypted, out...)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptByPrivateKey 使用RSA私钥解密数据
func DecryptByPrivateKey(encryptedData, privateKeyPEM string) (string, error) {
	// 获取私钥
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return "", err
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	// 分块解密
	var decryptedData []byte
	for i := 0; i < len(ciphertext); i += MaxDecryptBlock {
		end := i + MaxDecryptBlock
		if end > len(ciphertext) {
			end = len(ciphertext)
		}

		chunk := ciphertext[i:end]
		decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, chunk)
		if err != nil {
			return "", err
		}

		decryptedData = append(decryptedData, decryptedChunk...)
	}

	// URL解码
	decodedData, err := url.QueryUnescape(string(decryptedData))
	if err != nil {
		return "", err
	}

	return decodedData, nil
}

// parsePrivateKey 解析PEM格式的私钥
func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	// 解码Base64格式的私钥
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	// 解析PKCS#8格式的私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	// 类型断言
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA private key")
	}

	return rsaPrivateKey, nil
}

// GenerateRSAKeyPair 生成RSA密钥对（辅助函数）
func GenerateRSAKeyPair(bits int) (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}

	// 生成私钥的PKCS#8格式
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	// 生成公钥
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	privateKeyPEM := base64.StdEncoding.EncodeToString(privateKeyBytes)
	publicKeyPEM := base64.StdEncoding.EncodeToString(publicKeyBytes)

	return privateKeyPEM, publicKeyPEM, nil
}
