package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// 用于表示log记录的类型
type Level int8

const (
	LevelInfo Level = iota // 这样定义出来的类型全都会有String()方法用户返回相应的字符串
	LevelError
	LevelFatal
	LevelOff
)

// 根据传入的信息层级返回相应的字符串
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "Error"
	case LevelFatal:
		return "Fatal"
	default:
		return ""
	}
}

// 使用结构体存储logger的输出流 层级 锁 都是未导出的状态
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// 根据传入的输出流与层级返回Logger
func New(out io.Writer, minLevel Level) *Logger {
	// 锁不需要初始化
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// 输出一般级别的日志
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

// 输出错误级别的日志
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

// 输出致命级别的日志
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	// 结束程序的运行
	os.Exit(1)
}

// 未导出的内部print方法 -> 类似于用print对out(io.Writer).Write()进行了重写
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// 如果严重等级低于当前logger内部的最低级直接返回
	if level < l.minLevel {
		return 0, nil
	}
	// 定义结构体存储JSON信息用于输出
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().Local().Format(time.RFC3339), // 将本地时间(原先是UTC)格式化为string
		Message:    message,
		Properties: properties,
	}
	// 如果当前需要进行记录的信息等级大于等于LevelError则添加debug信息
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}
	var line []byte
	// 尝试将信息转换为JSON
	line, err := json.Marshal(aux)
	if err != nil {
		// 若解析为JSON失败
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}
	// 防止log并发输出导致结果混在一起
	l.mu.Lock()
	defer l.mu.Unlock()
	// 调用io.Writer.Write方法进行写入(返回写入的字节数并判断与实际值是否匹配)添加换行符
	return l.out.Write(append(line, '\n'))
}

// 实现Write接口 io.Writer.Write()
func (l *Logger) Write(message []byte) (int, error) {
	return l.print(LevelError, string(message), nil)
}
