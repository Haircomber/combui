//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/base64"
	"github.com/siongui/godom/wasm"
	"strings"
	"syscall/js"
)

type JQuery = wasm.Value

func Value(item JQuery) string {
	return item.Get("value").String()
}
func Checked(item JQuery) bool {
	return item.Get("checked").Bool()
}
func Name(item JQuery) string {
	return item.Get("name").String()
}
func ClientWidth(item JQuery) int {
	return item.Get("clientWidth").Int()
}
func WheelDeltaY(item JQuery) int {
	return item.Get("wheelDeltaY").Int()
}
func AppendChild(item JQuery, child, val string) {
	ch := wasm.Document.Call("createElement", child)
	ch.Set("innerHTML", val)
	item.Call("appendChild", ch)
}
func AppendChildId(item JQuery, child, val string) {
	ch := wasm.Document.Call("createElement", child)
	ch.Set("id", val)
	item.Call("appendChild", ch)
}
func AppendChildClass(item JQuery, child, val, class string) {
	ch := wasm.Document.Call("createElement", child)
	ch.Set("innerHTML", val)
	if len(class) > 0 {
		AddClass(wasm.Value{ch}, class)
	}
	item.Call("appendChild", ch)
}
func GetTarget(args []js.Value, target string) JQuery {
	return JQuery(wasm.Value{wasm.Value{args[0]}.Get(target)})
}

func AddWindowEventListener(event, target string, cb func(this, target JQuery)) {
	Window.Call("addEventListener", event, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb(JQuery{this}, GetTarget(args, target))
		return nil
	}))
}
func AddEventListener(item wasm.Value, event, target string, cb func(this, target JQuery)) {
	item.Call("addEventListener", event, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb(JQuery{this}, GetTarget(args, target))
		return nil
	}))
}

func GetAsCount(item wasm.Value) int {
	return item.Get("items").Get("length").Int()
}
func GetFileCount(item wasm.Value) int {
	return item.Get("files").Get("length").Int()
}

func GetAsString(item wasm.Value, id, target string, cb func(this JQuery, target string)) bool {
	if item.Get("items").Get("length").String() == id {
		return false
	}
	var thing = item.Get("items").Get(id)
	if thing.IsUndefined() {
		return false
	}
	if thing.Get("kind").String() == "string" {
		thing.Call("getAsString", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			cb(JQuery{this}, args[0].String())
			return nil
		}))
	} else if thing.Get("kind").String() == "file" {
		thing.Call("getAsFile").Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			cb(JQuery{this}, args[0].String())
			return nil
		}))
	} else {
		Alert(thing.Get("kind").String())
		return false
	}
	return true
}
func GetFileString(item wasm.Value, id, target string, cb func(this JQuery, target string)) bool {
	if item.Get("files").Get("length").String() == id {
		return false
	}
	var thing = item.Get("files").Get(id)
	if thing.IsUndefined() {
		return false
	}

	jsFileReader := js.Global().Get("FileReader")
	fr := jsFileReader.New()

	fr.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb(JQuery{this}, this.Get("result").String())
		return nil
	}))
	fr.Call("readAsText", thing)

	return true
}
func GetFileBase64(item wasm.Value, id, target string, cb func(this JQuery, target string)) bool {
	if item.Get("files").Get("length").String() == id {
		return false
	}
	var thing = item.Get("files").Get(id)
	if thing.IsUndefined() {
		return false
	}

	jsFileReader := js.Global().Get("FileReader")
	fr := jsFileReader.New()

	fr.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb(JQuery{this}, this.Get("result").String())
		return nil
	}))
	fr.Call("readAsDataURL", thing)

	return true
}
func Undisplay(item JQuery) {
	Display(item, "none")
}
func DisplayBlock(item JQuery) {
	Display(item, "block")
}
func DisplayInline(item JQuery) {
	Display(item, "inline")
}
func HideX(item JQuery) {
	Visibility(item, "hidden")
}
func ShowX(item JQuery) {
	Visibility(item, "visible")
}
func Display(item JQuery, display string) {
	item.Get("style").Call("setProperty", "display", display)
}
func Visibility(item JQuery, visibility string) {
	item.Get("style").Call("setProperty", "visibility", visibility)
}
func Width(item JQuery, width string) {
	item.Get("style").Call("setProperty", "width", width)
}
func SetValue(item JQuery, value string) {
	item.Set("value", value)
}
func SetSrc(item JQuery, value string) {
	item.Set("src", value)
}
func AddClass(item JQuery, value string) {
	item.Get("classList").Call("add", value)
}
func RemoveClass(item JQuery, value string) {
	item.Get("classList").Call("remove", value)
}
var jQuery = wasm.Document.QuerySelector //for convenience

func DocumentOrigin() string {
	return Document.Get("location").Get("origin").String()
}
func DocumentHash() string {
	return Document.Get("location").Get("hash").String()
}

func DownloadBuggedOnMobile(data, name string) {
	a := wasm.Document.Call("createElement", "a")
	a.Set("download", name)
	a.Set("rel", "noopener")
	a.Set("href", "data:text/plain;charset=utf-8,"+strings.Replace(data, "\r\n", `%0d%0a`, -1))
	a.Call("click")
	a.Call("remove")
}
func DownloadBase64(data, name string) {
	a := wasm.Document.Call("createElement", "a")
	a.Set("download", name)
	a.Set("rel", "noopener")
	a.Set("href", "data:text/plain;charset=utf-8;base64,"+base64.StdEncoding.EncodeToString([]byte(data)))
	a.Call("click")
	a.Call("remove")
}

func ReplaceState(h JQuery, item map[string]interface{}, str1, str2 string) {
	h.Call("replaceState", item, str1, str2)
}

var History = JQuery{js.Global().Get("history")}

func SetTimeout(cb func(this JQuery), time int) {
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb(JQuery{this})
		return nil
	}), time)
}

var Alert = wasm.Alert
var Document = wasm.Document
var Window = wasm.Window
