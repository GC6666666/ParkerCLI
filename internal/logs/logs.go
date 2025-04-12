package logs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/parker/ParkerCli/pkg/logger"
)

// LogStream 表示日志流
type LogStream struct {
	Lines    chan string
	File     string
	stopChan chan struct{}
}

// Close 停止日志流
func (l *LogStream) Close() {
	close(l.stopChan)
}

// LogsManager 日志管理器接口
type LogsManager interface {
	// TailLogs 追踪日志文件
	TailLogs(file string, lines int, interval time.Duration) (*LogStream, error)
	// FilterLogs 过滤日志
	FilterLogs(file string, keyword string, ignoreCase bool, useRegex bool) ([]string, int, error)
	// ReadLastLines 读取最后几行
	ReadLastLines(file string, lines int) ([]string, error)
}

// StandardLogsManager 标准日志管理器实现
type StandardLogsManager struct{}

// NewStandardLogsManager 创建新的标准日志管理器
func NewStandardLogsManager() *StandardLogsManager {
	return &StandardLogsManager{}
}

// TailLogs 追踪日志文件
func (m *StandardLogsManager) TailLogs(file string, lines int, interval time.Duration) (*LogStream, error) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("日志文件不存在: %s", file)
		}
		return nil, fmt.Errorf("获取文件信息失败: %s", err)
	}

	// 创建日志流
	logStream := &LogStream{
		Lines:    make(chan string, 100), // 使用缓冲通道，避免阻塞
		File:     file,
		stopChan: make(chan struct{}),
	}

	// 读取最后几行
	lastLines, err := m.ReadLastLines(file, lines)
	if err != nil {
		return nil, fmt.Errorf("读取最后几行失败: %s", err)
	}

	// 记录初始文件大小
	lastSize := fileInfo.Size()

	// 启动追踪协程
	go func() {
		defer close(logStream.Lines)

		// 首先输出最后几行
		for _, line := range lastLines {
			select {
			case logStream.Lines <- line:
			case <-logStream.stopChan:
				return
			}
		}

		// 定期检查文件变化
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 获取当前文件信息
				currentFileInfo, err := os.Stat(file)
				if err != nil {
					logger.Error("获取文件信息失败: %s", err)
					continue
				}

				currentSize := currentFileInfo.Size()

				// 文件大小变化检测 - 处理文件被截断的情况
				if currentSize < lastSize {
					logger.Info("文件大小减小，可能被截断，重新从头开始读取")
					lastSize = 0
				}

				// 如果文件大小变大，读取新内容
				if currentSize > lastSize {
					f, err := os.Open(file)
					if err != nil {
						logger.Error("打开文件失败: %s", err)
						continue
					}

					// 定位到上次读取的位置
					_, err = f.Seek(lastSize, 0)
					if err != nil {
						logger.Error("定位文件位置失败: %s", err)
						f.Close()
						continue
					}

					// 读取新行
					scanner := bufio.NewScanner(f)
					hasNewContent := false

					for scanner.Scan() {
						hasNewContent = true
						text := scanner.Text()

						select {
						case logStream.Lines <- text:
						case <-logStream.stopChan:
							f.Close()
							return
						}
					}

					// 检查扫描错误
					if err := scanner.Err(); err != nil {
						logger.Error("扫描文件失败: %s", err)
					}

					// 只有在成功读取内容后才更新文件位置
					if hasNewContent {
						// 确保我们获取精确的当前位置
						currentPos, err := f.Seek(0, io.SeekCurrent)
						if err != nil {
							logger.Error("获取当前文件位置失败: %s", err)
							lastSize = currentSize // 回退到使用文件大小
						} else {
							lastSize = currentPos
						}

						logger.Debug("文件大小从 %d 增加到 %d，已读取到位置 %d", lastSize, currentSize, currentPos)
					}

					f.Close()
				}
			case <-logStream.stopChan:
				return
			}
		}
	}()

	return logStream, nil
}

// FilterLogs 过滤日志
func (m *StandardLogsManager) FilterLogs(file string, keyword string, ignoreCase bool, useRegex bool) ([]string, int, error) {
	// 检查文件是否存在
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, 0, fmt.Errorf("日志文件不存在: %s", file)
	}

	// 准备正则表达式
	var re *regexp.Regexp
	if useRegex {
		var err error
		if ignoreCase {
			re, err = regexp.Compile("(?i)" + keyword)
		} else {
			re, err = regexp.Compile(keyword)
		}
		if err != nil {
			return nil, 0, fmt.Errorf("无效的正则表达式: %s", err)
		}
	} else if ignoreCase {
		keyword = strings.ToLower(keyword)
	}

	// 打开文件
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, fmt.Errorf("打开文件失败: %s", err)
	}
	defer f.Close()

	// 读取并过滤日志
	var matches []string
	lineCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		var match bool
		if useRegex {
			match = re.MatchString(line)
		} else if ignoreCase {
			match = strings.Contains(strings.ToLower(line), keyword)
		} else {
			match = strings.Contains(line, keyword)
		}

		if match {
			matches = append(matches, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("读取文件失败: %s", err)
	}

	return matches, lineCount, nil
}

// ReadLastLines 读取文件最后几行
func (m *StandardLogsManager) ReadLastLines(file string, lines int) ([]string, error) {
	// 打开文件
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %s", err)
	}
	defer f.Close()

	// 读取所有行
	var allLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %s", err)
	}

	// 返回最后几行
	if len(allLines) <= lines {
		return allLines, nil
	}
	return allLines[len(allLines)-lines:], nil
}

// FormatFilterResultsSummary 格式化过滤结果摘要
func FormatFilterResultsSummary(matches []string, lineCount int, keyword string, filePath string) string {
	summary := fmt.Sprintf("\n匹配: %d 行 / 总计: %d 行 (关键字: '%s')\n",
		len(matches), lineCount, keyword)
	summary += fmt.Sprintf("文件: %s\n", filePath)

	return summary
}
