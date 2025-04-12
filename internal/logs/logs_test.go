package logs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 创建测试日志文件
func createTestLogFile(t *testing.T, content string) string {
	tmpDir, err := ioutil.TempDir("", "parkerlogtest")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	tmpFile := filepath.Join(tmpDir, "test.log")
	if err := ioutil.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试日志文件失败: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpFile
}

// 测试过滤日志功能
func TestFilterLogs(t *testing.T) {
	// 创建测试日志
	logContent := `2023-10-12 10:00:00 INFO: 系统启动成功
2023-10-12 10:01:00 WARN: 内存使用超过80%
2023-10-12 10:02:00 ERROR: 无法连接数据库
2023-10-12 10:03:00 INFO: 用户登录 (ID: 12345)
2023-10-12 10:04:00 INFO: API请求: GET /users
2023-10-12 10:05:00 DEBUG: 查询执行时间: 213ms`

	testFile := createTestLogFile(t, logContent)

	// 创建日志管理器
	manager := NewStandardLogsManager()

	// 测试基本关键字过滤
	t.Run("Basic Keyword Filter", func(t *testing.T) {
		matches, total, err := manager.FilterLogs(testFile, "INFO", false, false)
		if err != nil {
			t.Fatalf("过滤日志失败: %v", err)
		}

		if total != 6 {
			t.Errorf("总行数应为6，实际为%d", total)
		}

		if len(matches) != 3 {
			t.Errorf("应匹配3行含有INFO的日志，实际匹配%d行", len(matches))
		}
	})

	// 测试忽略大小写选项
	t.Run("Case Insensitive Filter", func(t *testing.T) {
		matches, _, err := manager.FilterLogs(testFile, "info", true, false)
		if err != nil {
			t.Fatalf("过滤日志失败: %v", err)
		}

		if len(matches) != 3 {
			t.Errorf("忽略大小写时应匹配3行含有info的日志，实际匹配%d行", len(matches))
		}
	})

	// 测试正则表达式过滤
	t.Run("Regex Filter", func(t *testing.T) {
		matches, _, err := manager.FilterLogs(testFile, "\\d{5}", false, true)
		if err != nil {
			t.Fatalf("过滤日志失败: %v", err)
		}

		if len(matches) != 1 {
			t.Errorf("应匹配1行含有5位数字的日志，实际匹配%d行", len(matches))
		}
	})
}

// 测试读取最后几行功能
func TestReadLastLines(t *testing.T) {
	// 创建测试日志
	logContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10"
	testFile := createTestLogFile(t, logContent)

	// 创建日志管理器
	manager := NewStandardLogsManager()

	// 测试读取最后3行
	t.Run("Read Last 3 Lines", func(t *testing.T) {
		lines, err := manager.ReadLastLines(testFile, 3)
		if err != nil {
			t.Fatalf("读取最后几行失败: %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("应该读取3行，实际读取%d行", len(lines))
		}

		if lines[0] != "Line 8" || lines[1] != "Line 9" || lines[2] != "Line 10" {
			t.Errorf("读取的行内容不正确: %v", lines)
		}
	})

	// 测试请求行数超过文件总行数
	t.Run("Request More Lines Than File Has", func(t *testing.T) {
		lines, err := manager.ReadLastLines(testFile, 20)
		if err != nil {
			t.Fatalf("读取最后几行失败: %v", err)
		}

		if len(lines) != 10 {
			t.Errorf("应该读取全部10行，实际读取%d行", len(lines))
		}
	})
}

// 测试日志追踪功能
func TestTailLogs(t *testing.T) {
	// 测试简单初始读取
	t.Run("Initial Reading", func(t *testing.T) {
		// 创建测试数据
		logContent := "Line 1\nLine 2\nLine 3"
		testFile := createTestLogFile(t, logContent)
		t.Logf("创建测试文件: %s", testFile)

		// 创建日志管理器
		manager := NewStandardLogsManager()

		// 启动日志跟踪
		logStream, err := manager.TailLogs(testFile, 2, 50*time.Millisecond)
		if err != nil {
			t.Fatalf("启动日志追踪失败: %v", err)
		}
		defer logStream.Close()

		// 读取并验证初始行
		line1 := <-logStream.Lines
		line2 := <-logStream.Lines

		if line1 != "Line 2" {
			t.Errorf("第一行内容不符，预期: 'Line 2', 实际: '%s'", line1)
		}
		if line2 != "Line 3" {
			t.Errorf("第二行内容不符，预期: 'Line 3', 实际: '%s'", line2)
		}
	})

	// 创建一个非常简单的测试，只检查文件变化检测
	t.Run("Simple Change Detection", func(t *testing.T) {
		// 创建一个简单的测试文件
		initialContent := "Initial line"
		testFile := createTestLogFile(t, initialContent)

		// 创建日志管理器
		manager := NewStandardLogsManager()

		// 启动日志跟踪，不读取历史内容
		logStream, err := manager.TailLogs(testFile, 0, 100*time.Millisecond)
		if err != nil {
			t.Fatalf("启动日志追踪失败: %v", err)
		}
		defer logStream.Close()

		// 确保初始化完成
		time.Sleep(200 * time.Millisecond)

		// 添加单行新内容
		newLine := "New line content"
		t.Logf("添加新行: %s", newLine)

		f, err := os.OpenFile(testFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("打开文件失败: %v", err)
		}

		_, err = f.WriteString("\n" + newLine)
		if err != nil {
			f.Close()
			t.Fatalf("写入文件失败: %v", err)
		}

		f.Sync()
		f.Close()

		// 等待足够长的时间以确保检测到变化
		time.Sleep(500 * time.Millisecond)

		// 使用超时来验证是否收到更新
		select {
		case receivedLine := <-logStream.Lines:
			if receivedLine == newLine {
				// 测试通过 - 收到了正确的行
				t.Logf("成功检测到新行: %s", receivedLine)
			} else {
				t.Errorf("接收到的行内容不正确，预期: '%s', 实际: '%s'", newLine, receivedLine)
			}
		case <-time.After(3 * time.Second):
			t.Errorf("未能检测到文件变化，超时")
		}
	})
}

// 辅助函数：读取指定数量的行，带超时
func readLinesWithTimeout(t *testing.T, logStream *LogStream, count int, timeout time.Duration) []string {
	t.Helper()
	var lines []string
	timeoutChan := time.After(timeout)

	for i := 0; i < count; i++ {
		select {
		case line := <-logStream.Lines:
			t.Logf("读取到行: %s", line)
			lines = append(lines, line)
		case <-timeoutChan:
			t.Logf("超时等待行，已读取 %d/%d 行: %v", len(lines), count, lines)
			return lines
		}
	}

	return lines
}

// 辅助函数：向文件追加内容
func appendToFile(filepath, content string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	// 确保内容写入磁盘
	return f.Sync()
}

// 测试两个字符串数组是否包含相同的元素（顺序不重要）
func testEqualStringSlices(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}

	// 创建一个映射表来计数
	counts := make(map[string]int)

	// 统计预期行
	for _, line := range expected {
		counts[line]++
	}

	// 验证实际行
	for _, line := range actual {
		if counts[line] <= 0 {
			return false // 找到一个不在预期中的行
		}
		counts[line]--
	}

	// 所有计数应该为0
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}

	return true
}
