package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

var (
	key   = "vIUiXqgAAShtVymZCkra1sGCydma8OwhktMB5rJvRFc="
	token = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ0ZXN0LXVzZXIifQ.otjBRz4_DaixJO6FnOZuqsuw3veEq9gl10niWURVVcU"
)

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func main() {
	// 使用与 Java 示例相同的密钥
	secretKey := []byte(key)

	// 创建头部
	header := map[string]interface{}{
		"alg": "HS256",
	}

	// 创建负载
	claims := jwt.MapClaims{
		"sub": "test-user",
	}

	// 编码头部
	headerJSON, err := json.Marshal(header)
	if err != nil {
		fmt.Println("Error marshaling header:", err)
		return
	}
	headerEncoded := base64URLEncode(headerJSON)

	// 编码负载
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		fmt.Println("Error marshaling claims:", err)
		return
	}
	claimsEncoded := base64URLEncode(claimsJSON)

	// 拼接头部和负载
	unsignedToken := headerEncoded + "." + claimsEncoded

	// 使用密钥对令牌签名
	signingMethod := jwt.SigningMethodHS256
	signature, err := signingMethod.Sign(unsignedToken, secretKey)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return
	}

	// 拼接签名
	token := unsignedToken + "." + signature

	fmt.Println("Generated JWT:", token)

	// 验证 JWT
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		fmt.Println("Error parsing token:", err)
		return
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		fmt.Println("JWT is valid")
		fmt.Println("Subject:", claims["sub"])
	} else {
		fmt.Println("JWT is invalid")
	}
}
