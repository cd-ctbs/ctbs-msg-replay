package main

import (
	_ "bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	_ "golang.org/x/text/transform"
	"io/ioutil"
	"strings"
)

// DecryptFile 使用DES ECB模式解密文件
func DecryptFile(encryptedFilePath, encKey string) (string, error) {
	// 读取加密文件内容
	encryptedText, err := ioutil.ReadFile(encryptedFilePath)
	if err != nil {
		return "", fmt.Errorf("读取加密文件失败: %v", err)
	}

	encryptedTextStr := string(encryptedText)
	if encryptedTextStr == "" {
		return "", fmt.Errorf("加密文件内容为空")
	}

	// 将十六进制字符串转换为字节数组
	encryptedMessageBytes, err := hex.DecodeString(encryptedTextStr)
	if err != nil {
		return "", fmt.Errorf("解码十六进制字符串失败: %v", err)
	}

	// 创建DES密钥
	passwordBytes := []byte(encKey)
	desKey := make([]byte, 8)
	copy(desKey, passwordBytes)
	if len(passwordBytes) > 8 {
		copy(desKey, passwordBytes[:8])
	}

	// 创建DES解密器
	block, err := des.NewCipher(desKey)
	if err != nil {
		return "", fmt.Errorf("创建DES密码器失败: %v", err)
	}

	// 使用ECB模式解密
	decrypted, err := decryptECB(block, encryptedMessageBytes)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	// 转换为GBK编码的字符串
	gbkDecoder := simplifiedchinese.GBK.NewDecoder()
	decryptedStr, err := gbkDecoder.String(string(decrypted))
	if err != nil {
		return "", fmt.Errorf("GBK解码失败: %v", err)
	}

	return strings.TrimSpace(decryptedStr), nil
}

// decryptECB 实现ECB模式解密
func decryptECB(block cipher.Block, ciphertext []byte) ([]byte, error) {
	blockSize := block.BlockSize()

	// 确保密文长度是块大小的倍数
	if len(ciphertext)%blockSize != 0 {
		return nil, fmt.Errorf("密文长度不是块大小的倍数")
	}

	decrypted := make([]byte, len(ciphertext))

	// 对每个块进行解密
	for i := 0; i < len(ciphertext); i += blockSize {
		block.Decrypt(decrypted[i:i+blockSize], ciphertext[i:i+blockSize])
	}

	// 去除PKCS5填充
	return pkcs5Unpad(decrypted)
}

// pkcs5Unpad 去除PKCS5填充
func pkcs5Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("数据长度为0")
	}

	// 获取最后一个字节作为填充长度
	unpadding := int(data[length-1])

	// 检查填充是否有效
	if unpadding > length {
		return nil, fmt.Errorf("无效的PKCS5填充")
	}

	// 返回去除填充后的数据
	return data[:(length - unpadding)], nil
}

// SaveFile 保存解密后的文件（对应C#中的SaveFile方法）
func SaveFile(text, fileName string) error {
	// 将文本转换为GBK编码
	gbkEncoder := simplifiedchinese.GBK.NewEncoder()
	gbkText, err := gbkEncoder.String(text)
	if err != nil {
		return fmt.Errorf("GBK编码失败: %v", err)
	}

	// 写入文件
	err = ioutil.WriteFile(fileName, []byte(gbkText), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}
