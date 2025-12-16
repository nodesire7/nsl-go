/**
 * 二维码工具
 * 生成二维码图片
 */
package utils

import (
	"bytes"
	"encoding/base64"
	"image/png"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCode 生成二维码并返回Base64字符串
func GenerateQRCode(url string, size int) (string, error) {
	if size <= 0 {
		size = 256
	}

	// 生成二维码
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return "", err
	}

	// 转换为PNG
	var buf bytes.Buffer
	img := qr.Image(size)
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	// 转换为Base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + base64Str, nil
}

