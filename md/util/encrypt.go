// 加密工具类
package util

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

type SignType string

const (
	MD5    SignType = "MD5"
	SHA1   SignType = "SHA1"
	SHA256 SignType = "SHA256"
	SHA512 SignType = "SHA512"
)

// MD5加密
func EncryptMD5(message []byte) string {
	hash := md5.New()
	hash.Write(message)
	sum := hash.Sum(nil)
	hashCode := hex.EncodeToString(sum)
	return hashCode
}

// SHA1加密
func EncryptSHA1(message []byte) string {
	hash := sha1.New()
	hash.Write(message)
	sum := hash.Sum(nil)
	hashCode := hex.EncodeToString(sum)
	return hashCode
}

// SHA256加密
func EncryptSHA256(message []byte) string {
	hash := sha256.New()
	hash.Write(message)
	sum := hash.Sum(nil)
	hashCode := hex.EncodeToString(sum)
	return hashCode
}

// SHA512加密
func EncryptSHA512(message []byte) string {
	hash := sha512.New()
	hash.Write(message)
	sum := hash.Sum(nil)
	hashCode := hex.EncodeToString(sum)
	return hashCode
}

// BASE64编码
func EncryptBASE64(message []byte) string {
	return base64.StdEncoding.EncodeToString(message)
}

// BASE64解码
func DecryptBASE64(message string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(message)
}

// 生成RSA密钥对
func GenerateRSAKey(bits int, isPKCS8 bool) (string, string, error) {
	if bits < 512 || bits > 4096 {
		return "", "", errors.New("密钥位数需在512-4096之间")
	}
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", "", err
	}
	var privateDer []byte
	if isPKCS8 {
		privateDer, err = x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return "", "", err
		}
	} else {
		privateDer = x509.MarshalPKCS1PrivateKey(privateKey)
	}
	// 生成公钥
	publicDer, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	return EncryptBASE64(privateDer), EncryptBASE64(publicDer), nil
}

// RSA公钥加密
func EncryptRSA(message, publicKey string) (string, error) {
	key, err := DecryptBASE64(publicKey)
	if err != nil {
		return "", err
	}
	pubKey, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey.(*rsa.PublicKey), []byte(message))
	if err != nil {
		return "", err
	}
	return EncryptBASE64(encryptedData), nil
}

// RSA私钥解密
func DecryptRSA(message, privateKey string, isPKCS8 bool) (string, error) {
	messageBytes, err := DecryptBASE64(message)
	if err != nil {
		return "", err
	}
	key, err := DecryptBASE64(privateKey)
	if err != nil {
		return "", err
	}
	var priKey interface{}
	if isPKCS8 {
		priKey, err = x509.ParsePKCS8PrivateKey(key)
	} else {
		priKey, err = x509.ParsePKCS1PrivateKey(key)
	}
	if err != nil {
		return "", err
	}
	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, priKey.(*rsa.PrivateKey), messageBytes)
	if err != nil {
		return "", err
	}
	return string(decryptedData), nil
}

// RSA私钥签名
func SignRSA(message, privateKey string, signType SignType, isPKCS8 bool) (string, error) {
	key, err := DecryptBASE64(privateKey)
	if err != nil {
		return "", err
	}
	var priKey interface{}
	if isPKCS8 {
		priKey, err = x509.ParsePKCS8PrivateKey(key)
	} else {
		priKey, err = x509.ParsePKCS1PrivateKey(key)
	}
	if err != nil {
		return "", err
	}
	var signature []byte
	switch signType {
	case MD5:
		h := md5.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		signature, err = rsa.SignPKCS1v15(rand.Reader, priKey.(*rsa.PrivateKey), crypto.MD5, hash)
	case SHA1:
		h := sha1.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		signature, err = rsa.SignPKCS1v15(rand.Reader, priKey.(*rsa.PrivateKey), crypto.SHA1, hash)
	case SHA256:
		h := sha256.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		signature, err = rsa.SignPKCS1v15(rand.Reader, priKey.(*rsa.PrivateKey), crypto.SHA256, hash)
	case SHA512:
		h := sha512.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		signature, err = rsa.SignPKCS1v15(rand.Reader, priKey.(*rsa.PrivateKey), crypto.SHA512, hash)
	default:
		return "", errors.New("不支持的签名类型")
	}
	if err != nil {
		return "", err
	}
	return EncryptBASE64(signature), nil
}

// RSA公钥验签
func VerifyRSA(message, publicKey, sign string, signType SignType) error {
	signBytes, err := DecryptBASE64(sign)
	if err != nil {
		return err
	}
	key, err := DecryptBASE64(publicKey)
	if err != nil {
		return err
	}
	pubKey, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		return err
	}
	switch signType {
	case MD5:
		h := md5.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.MD5, hash, signBytes)
	case SHA1:
		h := sha1.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA1, hash, signBytes)
	case SHA256:
		h := sha256.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hash, signBytes)
	case SHA512:
		h := sha512.New()
		h.Write([]byte(message))
		hash := h.Sum(nil)
		err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA512, hash, signBytes)
	default:
		return errors.New("不支持的签名类型")
	}
	if err != nil {
		return err
	}
	return nil
}

// PKCS7填充
func pkcs7Padding(message []byte, blockSize int) []byte {
	padding := blockSize - len(message)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(message, padText...)
}

// PKCS7 unpadding with validation.
// Returns error if padding is invalid.
func pkcs7UnPadding(message []byte) ([]byte, error) {
	length := len(message)
	if length == 0 {
		return nil, errors.New("empty message")
	}
	unPadding := int(message[length-1])
	if unPadding == 0 || unPadding > length || unPadding > 16 {
		return nil, errors.New("invalid padding")
	}
	// Verify all padding bytes are correct
	for i := length - unPadding; i < length; i++ {
		if message[i] != byte(unPadding) {
			return nil, errors.New("invalid padding")
		}
	}
	return message[:(length - unPadding)], nil
}

// Pad key to valid AES key length (16, 24, or 32 bytes).
// Truncates if longer than 32 bytes.
func paddingKey(key string) []byte {
	keyByte := []byte(key)
	keyLen := len(keyByte)

	// Find target length
	targetLen := 16
	if keyLen > 16 {
		targetLen = 24
	}
	if keyLen > 24 {
		targetLen = 32
	}

	// Truncate or pad
	if keyLen > 32 {
		return keyByte[:32]
	}
	if keyLen < targetLen {
		padded := make([]byte, targetLen)
		copy(padded, keyByte)
		return padded
	}
	return keyByte
}

// AES encrypt using CBC mode with random IV.
// The IV is prepended to the ciphertext.
func EncryptAES(message, key string, isHex bool) (res string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("AES加密异常")
		}
	}()
	messageByte := []byte(message)
	keyByte := paddingKey(key)
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	messageByte = pkcs7Padding(messageByte, blockSize)

	// Generate random IV for each encryption
	iv := make([]byte, blockSize)
	if _, err = rand.Read(iv); err != nil {
		return "", err
	}

	blockMode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(messageByte))
	blockMode.CryptBlocks(ciphertext, messageByte)

	// Prepend IV to ciphertext
	result := append(iv, ciphertext...)

	if isHex {
		return hex.EncodeToString(result), nil
	}
	return EncryptBASE64(result), nil
}

// AES decrypt using CBC mode.
// Expects IV prepended to ciphertext.
// For backward compatibility, also supports legacy format without IV prefix.
func DecryptAES(message, key string, isHex bool) (res string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("AES解密异常")
		}
	}()
	var messageByte []byte
	if isHex {
		messageByte, err = hex.DecodeString(message)
	} else {
		messageByte, err = DecryptBASE64(message)
	}
	if err != nil {
		return "", err
	}

	keyByte := paddingKey(key)
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()

	// Try new format first (IV prepended)
	if len(messageByte) >= blockSize*2 {
		iv := messageByte[:blockSize]
		ciphertext := messageByte[blockSize:]

		blockMode := cipher.NewCBCDecrypter(block, iv)
		plaintext := make([]byte, len(ciphertext))
		blockMode.CryptBlocks(plaintext, ciphertext)

		result, unpadErr := pkcs7UnPadding(plaintext)
		if unpadErr == nil {
			return string(result), nil
		}
	}

	// Fall back to legacy format (key as IV) for backward compatibility
	if len(messageByte) >= blockSize {
		blockMode := cipher.NewCBCDecrypter(block, keyByte[:blockSize])
		plaintext := make([]byte, len(messageByte))
		blockMode.CryptBlocks(plaintext, messageByte)

		result, unpadErr := pkcs7UnPadding(plaintext)
		if unpadErr == nil {
			return string(result), nil
		}
	}

	return "", errors.New("decryption failed")
}
