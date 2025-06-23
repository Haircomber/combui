package main

import "net/http"
import "fmt"
import "bytes"
import "encoding/json"

func GetJson(url string) (json string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("response status code: %d", resp.StatusCode)
		return
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}
	json = buf.String()
	return
}

func Call(server, action string, callback func(string)) {

	json, err := GetJson(server + "/" + action)
	if err != nil {
		callback(`{"Success": false}`)
		return
	}
	callback(json)
}

func Parse(str string) (dat *AppModelResponseUniversal) {
	for i, c := range str {
		if c == '{' {
			str = str[i:]
			break
		}
	}
	for i := len(str); i > 1; i-- {
		if str[i-1] == '}' {
			str = str[:i]
			break
		}
	}
	var data AppModelSuccess
	err := json.Unmarshal([]byte(str), &data)
	if err == nil && data.Success {
		dat = new(AppModelResponseUniversal)
		err = json.Unmarshal([]byte(str), dat)
		if err != nil {
			return nil
		}
		return dat
	}
	return nil
}
