package primer

// Logger 日志记录器接口
type Logger interface {
	// Debug 记录调试信息
	Debug(msg string, fields ...Field)
	// Info 记录信息
	Info(msg string, fields ...Field)
	// Error 记录错误
	Error(msg string, fields ...Field)
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建 int64 字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// ErrorField 创建错误字段
func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

// NopLogger 空日志记录器（不记录任何日志）
type NopLogger struct{}

func (n *NopLogger) Debug(msg string, fields ...Field) {}
func (n *NopLogger) Info(msg string, fields ...Field)  {}
func (n *NopLogger) Error(msg string, fields ...Field) {}
