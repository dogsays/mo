package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

type partString struct {
	raw string
}

func (p *partString) Output(args *Args) {
	args.Buffer.WriteString(p.raw)
}

func PartString(raw string) *partString {
	return &partString{raw: raw}
}

func PartKV(k, v any) *partString {
	return PartString(fmt.Sprintf("%v=%v", k, v))
}

////

type partSince struct {
	t time.Time
}

func (p *partSince) Output(args *Args) {
	now := time.Now()
	args.Buffer.WriteString(now.Sub(p.t).String())
	// 不支持并发
	p.t = now
}

func PartSince() *partSince {
	return &partSince{t: time.Now()}
}

////

type Level int8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type partLevel struct {
}

func PartLevel() *partLevel {
	return &partLevel{}
}

var lv2str = []string{
	LevelDebug: "DBG",
	LevelInfo:  "INF",
	LevelWarn:  "WAN",
	LevelError: "ERR",
}

func (p *partLevel) Output(args *Args) {
	args.Buffer.WriteString(lv2str[args.Level])
}

////

type partCaller struct {
	shortFile bool
}

func PartCaller(shortFile bool) *partCaller {
	return &partCaller{shortFile: shortFile}
}

func (p *partCaller) Output(args *Args) {

	fs := runtime.CallersFrames([]uintptr{args.PC})
	f, _ := fs.Next()

	if p.shortFile {
		f.File = filepath.Base(f.File)
	}

	args.Buffer.WriteString(f.File)
	args.Buffer.Write(':')
	args.Buffer.WriteInt(f.Line)
}

////

type partDateTime struct {
	layout string
}

func PartDateTime(layout string) *partDateTime {
	return &partDateTime{layout: layout}
}

func (p *partDateTime) Output(args *Args) {
	args.Buffer.buf = args.Time.AppendFormat(args.Buffer.buf, p.layout)
}

////

type partMessage struct {
}

func PartMessage() *partMessage {
	return &partMessage{}
}

func (p *partMessage) Output(args *Args) {
	args.Buffer.WriteString(args.Msg)
}

////
