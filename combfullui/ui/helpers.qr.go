package main

import "encoding/base64"
import "github.com/skip2/go-qrcode"

func QRCODE(str string) string {
	png, err := qrcode.Encode(str, qrcode.Highest, 256)
	if err != nil {
		return ""
	}
	base64Img := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64Img
}