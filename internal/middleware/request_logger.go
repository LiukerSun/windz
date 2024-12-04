package middleware

import (
	"backend/pkg/config"
	"backend/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用请求日志
		if !config.GetBool("log.request_log") {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil && method != "GET" {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器以捕获响应体
		blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算请求处理时间
		duration := time.Since(start).Seconds()

		// 构建日志字段
		fields := map[string]interface{}{
			"method": method,
			"path":   path,
		}

		if query != "" {
			fields["query"] = query
		}

		fields["status"] = c.Writer.Status()
		fields["ip"] = c.ClientIP()
		fields["latency"] = fmt.Sprintf("%.3fs", duration)

		ua := c.Request.UserAgent()
		if ua != "" {
			fields["user_agent"] = ua
		}

		// 根据状态码选择日志级别
		logEntry := logger.WithFields(fields)
		status := c.Writer.Status()
		msg := fmt.Sprintf("%s %s", method, path)

		if status >= 500 {
			logEntry.Error(msg)
		} else if status >= 400 {
			logEntry.Warn(msg)
		} else {
			logEntry.Info(msg)
		}
	}
}

// getRequestHeaders 获取请求头
func getRequestHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		// 跳过敏感头部信息
		if strings.EqualFold(k, "Authorization") || strings.EqualFold(k, "Cookie") {
			continue
		}
		headers[k] = strings.Join(v, ";")
	}
	return headers
}

// getResponseHeaders 获取响应头
func getResponseHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for k, v := range c.Writer.Header() {
		headers[k] = strings.Join(v, ";")
	}
	return headers
}

// isJSON 检查字符串是否是JSON格式
func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// prettyJSON 格式化JSON字符串
func prettyJSON(str string) string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(str), "", "  "); err != nil {
		return str
	}
	return pretty.String()
}
