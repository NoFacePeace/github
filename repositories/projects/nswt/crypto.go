package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	safePrefix     = "V02_"
	randomAlphabet = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
)

var aesIV = []byte("0000000000000000")

type KeyPaths struct {
	TransportPrivate string `json:"transportPrivate"`
	TransportPublic  string `json:"transportPublic"`
	SafePrivate      string `json:"safePrivate"`
	SafePublic       string `json:"safePublic"`
}

type SafePlaintext struct {
	Header       string `json:"header"`
	Timestamp    int64  `json:"timestamp"`
	PasswordHash string `json:"passwordHash"`
	Nonce        string `json:"nonce"`
	Random       string `json:"random"`
}

type TransportEnvelope struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	AESKey  string            `json:"-"`
	JSON    string            `json:"-"`
}

type Position struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Location struct {
	Nation   string   `json:"nation,omitempty"`
	Adcode   string   `json:"adcode,omitempty"`
	Province string   `json:"province"`
	City     string   `json:"city"`
	CityCode string   `json:"citycode"`
	District string   `json:"district"`
	Township string   `json:"township"`
	Address  string   `json:"address"`
	Position Position `json:"position"`
	Decode   int      `json:"decode"`
}

type AsyncRushPayload struct {
	ID          string   `json:"id"`
	TemplateID  string   `json:"templateid"`
	CollectType int      `json:"collecttype"`
	BatchID     string   `json:"batchid"`
	BatchCode   string   `json:"batchcode"`
	CipherCode  string   `json:"ciphercode"`
	SafeSalt    string   `json:"safesalt"`
	Ciphertext  string   `json:"ciphertext"`
	Mobile      string   `json:"mobile"`
	IsSafeKey   int      `json:"issafekey"`
	T           int      `json:"t"`
	Channel     string   `json:"channel"`
	Position    Position `json:"position"`
	Location    Location `json:"location"`
}

type DecryptedTransport struct {
	Payload   AsyncRushPayload
	Timestamp int64
	AESKey    string
	JSON      string
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runOpenSSL(input []byte, args ...string) ([]byte, error) {
	command := exec.Command("openssl", args...)
	command.Stdin = bytes.NewReader(input)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("OpenSSL 执行失败: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	return stdout.Bytes(), nil
}

func ensureTestKeys(keysDir string) (KeyPaths, error) {
	keys := KeyPaths{
		TransportPrivate: filepath.Join(keysDir, "transport-private.pem"),
		TransportPublic:  filepath.Join(keysDir, "transport-public.pem"),
		SafePrivate:      filepath.Join(keysDir, "safe-private.pem"),
		SafePublic:       filepath.Join(keysDir, "safe-public.pem"),
	}
	if err := os.MkdirAll(keysDir, 0o755); err != nil {
		return KeyPaths{}, fmt.Errorf("创建密钥目录失败: %w", err)
	}

	if !fileExists(keys.TransportPrivate) || !fileExists(keys.TransportPublic) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return KeyPaths{}, fmt.Errorf("生成 RSA 密钥失败: %w", err)
		}
		privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return KeyPaths{}, fmt.Errorf("编码 RSA 私钥失败: %w", err)
		}
		publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return KeyPaths{}, fmt.Errorf("编码 RSA 公钥失败: %w", err)
		}
		if err := os.WriteFile(keys.TransportPrivate, pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateDER,
		}), 0o600); err != nil {
			return KeyPaths{}, err
		}
		if err := os.WriteFile(keys.TransportPublic, pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicDER,
		}), 0o644); err != nil {
			return KeyPaths{}, err
		}
	}

	if !fileExists(keys.SafePrivate) || !fileExists(keys.SafePublic) {
		if _, err := runOpenSSL(nil,
			"genpkey",
			"-algorithm", "EC",
			"-pkeyopt", "ec_paramgen_curve:SM2",
			"-out", keys.SafePrivate,
		); err != nil {
			return KeyPaths{}, err
		}
		if _, err := runOpenSSL(nil,
			"pkey",
			"-in", keys.SafePrivate,
			"-pubout",
			"-out", keys.SafePublic,
		); err != nil {
			return KeyPaths{}, err
		}
		if err := os.Chmod(keys.SafePrivate, 0o600); err != nil {
			return KeyPaths{}, err
		}
	}

	return keys, nil
}

func secureRandomString(length int) (string, error) {
	result := make([]byte, length)
	limit := byte(256 - (256 % len(randomAlphabet)))
	for index := range result {
		for {
			var value [1]byte
			if _, err := rand.Read(value[:]); err != nil {
				return "", err
			}
			if value[0] < limit {
				result[index] = randomAlphabet[int(value[0])%len(randomAlphabet)]
				break
			}
		}
	}
	return string(result), nil
}

func buildSafePlaintext(
	password, salt string,
	timestamp int64,
	nonce, randomValue string,
) ([]byte, error) {
	if nonce == "" {
		nonce = strconv.FormatInt(timestamp, 10)
	}
	if randomValue == "" {
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, fmt.Errorf("生成安全随机数失败: %w", err)
		}
		randomValue = fmt.Sprintf("%x", randomBytes)
	}

	body := fmt.Sprintf(
		"%d\x00%s\x00%s\x00%s",
		timestamp,
		md5Hex(password+salt),
		nonce,
		randomValue,
	)
	return append([]byte{0x00, 0x00}, []byte(body)...), nil
}

func parseSafePlaintext(plaintext []byte) (SafePlaintext, error) {
	if len(plaintext) < 3 || plaintext[0] != 0x00 || plaintext[1] != 0x00 {
		return SafePlaintext{}, fmt.Errorf("safe-password header 不是 custom hash (0x00 0x00)")
	}
	remaining := plaintext[2:]
	fields := make([]string, 0, 3)
	for range 3 {
		index := bytes.IndexByte(remaining, 0x00)
		if index < 0 {
			return SafePlaintext{}, fmt.Errorf("safe-password 明文字段不完整")
		}
		fields = append(fields, string(remaining[:index]))
		remaining = remaining[index+1:]
	}
	timestamp, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return SafePlaintext{}, fmt.Errorf("safe-password timestamp 无效: %w", err)
	}
	return SafePlaintext{
		Header:       "0000",
		Timestamp:    timestamp,
		PasswordHash: fields[1],
		Nonce:        fields[2],
		Random:       string(remaining),
	}, nil
}

func encryptSafePassword(
	password, salt string,
	timestamp int64,
	nonce, randomValue, publicKeyPath string,
) (string, error) {
	plaintext, err := buildSafePlaintext(password, salt, timestamp, nonce, randomValue)
	if err != nil {
		return "", err
	}
	encrypted, err := runOpenSSL(
		plaintext,
		"pkeyutl",
		"-encrypt",
		"-pubin",
		"-inkey", publicKeyPath,
	)
	if err != nil {
		return "", err
	}
	return safePrefix + base64.StdEncoding.EncodeToString(encrypted), nil
}

func decryptSafePassword(ciphertext, privateKeyPath string) (SafePlaintext, error) {
	if !strings.HasPrefix(ciphertext, safePrefix) {
		return SafePlaintext{}, fmt.Errorf("safe-password 密文缺少 %s 前缀", safePrefix)
	}
	encrypted, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, safePrefix))
	if err != nil {
		return SafePlaintext{}, fmt.Errorf("safe-password Base64 无效: %w", err)
	}
	plaintext, err := runOpenSSL(
		encrypted,
		"pkeyutl",
		"-decrypt",
		"-inkey", privateKeyPath,
	)
	if err != nil {
		return SafePlaintext{}, err
	}
	return parseSafePlaintext(plaintext)
}

func parseRSAPublicKey(path string) (*rsa.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("RSA 公钥不是 PEM")
	}
	value, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := value.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥不是 RSA")
	}
	return publicKey, nil
}

func parseRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, fmt.Errorf("RSA 私钥不是 PEM")
	}
	value, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := value.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥不是 RSA")
	}
	return privateKey, nil
}

func pkcs7Pad(input []byte, blockSize int) []byte {
	padding := blockSize - len(input)%blockSize
	return append(input, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs7Unpad(input []byte, blockSize int) ([]byte, error) {
	if len(input) == 0 || len(input)%blockSize != 0 {
		return nil, fmt.Errorf("AES 明文填充长度无效")
	}
	padding := int(input[len(input)-1])
	if padding == 0 || padding > blockSize || padding > len(input) {
		return nil, fmt.Errorf("AES PKCS7 填充无效")
	}
	for _, value := range input[len(input)-padding:] {
		if int(value) != padding {
			return nil, fmt.Errorf("AES PKCS7 填充不一致")
		}
	}
	return input[:len(input)-padding], nil
}

func encryptTransport(
	payload AsyncRushPayload,
	publicKeyPath string,
	timestamp int64,
) (TransportEnvelope, error) {
	publicKey, err := parseRSAPublicKey(publicKeyPath)
	if err != nil {
		return TransportEnvelope{}, err
	}
	jsonBytes, err := jsonMarshal(payload)
	if err != nil {
		return TransportEnvelope{}, err
	}
	aesKey, err := secureRandomString(16)
	if err != nil {
		return TransportEnvelope{}, err
	}
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return TransportEnvelope{}, err
	}
	padded := pkcs7Pad(jsonBytes, aes.BlockSize)
	encryptedBody := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, aesIV).CryptBlocks(encryptedBody, padded)

	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(aesKey))
	if err != nil {
		return TransportEnvelope{}, err
	}
	jsonBase64 := base64.StdEncoding.EncodeToString(jsonBytes)
	return TransportEnvelope{
		Body: base64.StdEncoding.EncodeToString(encryptedBody),
		Headers: map[string]string{
			"Content-Type":    "application/json; charset=utf-8",
			"X-APP-SN":        base64.StdEncoding.EncodeToString(encryptedKey),
			"X-APP-SIGN":      md5Hex(fmt.Sprintf("%d@%s@%s", timestamp, aesKey, jsonBase64)),
			"X-APP-TIMESTAMP": strconv.FormatInt(timestamp, 10),
		},
		AESKey: aesKey,
		JSON:   string(jsonBytes),
	}, nil
}

func decryptTransport(
	body string,
	headers map[string][]string,
	privateKeyPath string,
) (DecryptedTransport, error) {
	privateKey, err := parseRSAPrivateKey(privateKeyPath)
	if err != nil {
		return DecryptedTransport{}, err
	}
	encryptedKeyText := firstHeader(headers, "X-App-Sn")
	timestampText := firstHeader(headers, "X-App-Timestamp")
	actualSign := firstHeader(headers, "X-App-Sign")
	if encryptedKeyText == "" || timestampText == "" || actualSign == "" {
		return DecryptedTransport{}, fmt.Errorf("缺少 X-APP-SN、X-APP-SIGN 或 X-APP-TIMESTAMP")
	}
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyText)
	if err != nil {
		return DecryptedTransport{}, err
	}
	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedKey)
	if err != nil {
		return DecryptedTransport{}, fmt.Errorf("RSA 解密 AES key 失败: %w", err)
	}
	if len(aesKey) != 16 {
		return DecryptedTransport{}, fmt.Errorf("解出的 AES key 长度不是 16 字节")
	}
	encryptedBody, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return DecryptedTransport{}, fmt.Errorf("AES 请求体 Base64 无效: %w", err)
	}
	if len(encryptedBody) == 0 || len(encryptedBody)%aes.BlockSize != 0 {
		return DecryptedTransport{}, fmt.Errorf("AES 请求体长度无效")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return DecryptedTransport{}, err
	}
	plaintext := make([]byte, len(encryptedBody))
	cipher.NewCBCDecrypter(block, aesIV).CryptBlocks(plaintext, encryptedBody)
	plaintext, err = pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return DecryptedTransport{}, err
	}

	expectedSign := md5Hex(
		timestampText + "@" + string(aesKey) + "@" +
			base64.StdEncoding.EncodeToString(plaintext),
	)
	if strings.ToLower(actualSign) != expectedSign {
		return DecryptedTransport{}, fmt.Errorf(
			"X-APP-SIGN 校验失败: expected=%s, actual=%s",
			expectedSign,
			actualSign,
		)
	}
	timestamp, err := strconv.ParseInt(timestampText, 10, 64)
	if err != nil {
		return DecryptedTransport{}, fmt.Errorf("X-APP-TIMESTAMP 无效: %w", err)
	}
	var payload AsyncRushPayload
	if err := jsonUnmarshal(plaintext, &payload); err != nil {
		return DecryptedTransport{}, err
	}
	return DecryptedTransport{
		Payload:   payload,
		Timestamp: timestamp,
		AESKey:    string(aesKey),
		JSON:      string(plaintext),
	}, nil
}

func firstHeader(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func currentSeconds() int64 {
	return time.Now().Unix()
}

func currentMilliseconds() int64 {
	return time.Now().UnixMilli()
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func jsonMarshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func jsonUnmarshal(data []byte, value any) error {
	return json.Unmarshal(data, value)
}

func randomRequestID() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw)[:32], nil
}
