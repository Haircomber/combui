//go:build !js && !wasm
// +build !js,!wasm

package main

type JQuery string

func (JQuery) InnerHTML() string {
	return ""
}
func (JQuery) String() string {
	return ""
}
func (JQuery) Bool() bool {
	return false
}

func (JQuery) Call(_ string) {
}
func (JQuery) Set(_, _ string) {
}
func (v JQuery) Get(_ string) JQuery {
	return v
}
func (JQuery) SetInnerHTML(_ string) {
}

func Value(item JQuery) string {
	return item.Get("value").String()
}
func Checked(item JQuery) bool {
	return item.Get("checked").Bool()
}
func Name(item JQuery) string {
	return item.Get("name").String()
}
func AppendChild(item JQuery, child, val string) {
}
func AppendChildId(item JQuery, child, val string) {
}
func AppendChildClass(item JQuery, child, val, class string) {
}
func AddEventListener(item JQuery, event, target string, cb func(this, target JQuery)) {
}

func GetAsCount(item JQuery) int {
	return 0
}
func GetFileCount(item JQuery) int {
	return 0
}

func GetAsString(item JQuery, id, target string, cb func(this JQuery, target string)) bool {
	return false
}
func GetFileString(item JQuery, id, target string, cb func(this JQuery, target string)) bool {
	return false
}
func GetFileBase64(item JQuery, id, target string, cb func(this JQuery, target string)) bool {
	return false
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
}
func Visibility(item JQuery, visibility string) {
}
func Width(item JQuery, width string) {
}
func SetValue(item JQuery, value string) {
}
func SetSrc(item JQuery, value string) {
}
func AddClass(item JQuery, value string) {
}
func RemoveClass(item JQuery, value string) {
}
var jQuery = func(string) JQuery {
	return JQuery("")
}

func DocumentOrigin() string {
	return ""
}

func DocumentHash() string {
	return ""
}

func DownloadBuggedOnMobile(data, name string) {
}
func DownloadBase64(data, name string) {
}

func ReplaceState(h JQuery, item map[string]interface{}, str1, str2 string) {
}

var History JQuery

func SetTimeout(cb func(this JQuery), time int) {
}

var Alert JQuery
var Document JQuery
var Window JQuery
