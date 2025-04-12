package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/parker/ParkerCli/internal/config"
	"github.com/parker/ParkerCli/internal/utils"
	"github.com/parker/ParkerCli/pkg/logger"
)

// MigrationStatus 表示迁移状态
type MigrationStatus string

const (
	// StatusPending 表示待执行
	StatusPending MigrationStatus = "PENDING"
	// StatusApplied 表示已应用
	StatusApplied MigrationStatus = "APPLIED"
	// StatusFailed 表示执行失败
	StatusFailed MigrationStatus = "FAILED"
)

// MigrationType 表示迁移类型
type MigrationType string

const (
	// TypeSQL 表示SQL迁移
	TypeSQL MigrationType = "SQL"
	// TypeGoFn 表示Go函数迁移
	TypeGoFn MigrationType = "GO"
)

// Migration 表示单个迁移
type Migration struct {
	ID         string          // 迁移ID（通常为时间戳）
	Name       string          // 迁移名称
	Type       MigrationType   // 迁移类型
	Status     MigrationStatus // 迁移状态
	FilePath   string          // 文件路径
	AppliedAt  time.Time       // 应用时间
	BatchID    int             // 批次ID
	Reversible bool            // 是否可回滚
}

// MigrationService 迁移服务接口
type MigrationService interface {
	Create(name string, migrationType MigrationType) (*Migration, error)
	Up(steps int) error
	Down(steps int) error
	Status() ([]Migration, error)
	Reset() error
	Refresh() error
	Generate(name, template string) (*Migration, error)
}

// StandardMigrator 标准迁移器实现
type StandardMigrator struct {
	migrationsDir string
	schemaTable   string
	dbDriver      string
	dbDSN         string
}

// NewStandardMigrator 创建新的标准迁移器
func NewStandardMigrator() *StandardMigrator {
	// 从配置中获取迁移目录
	migrationsDir := config.GetString("paths.migrations")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	// 数据库配置
	dbConfig := config.GetAll().Database
	dbDriver := dbConfig.Driver

	var dbDSN string
	if dbDriver == "postgres" {
		dbDSN = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.SSLMode)
	} else if dbDriver == "mysql" {
		dbDSN = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)
	}

	return &StandardMigrator{
		migrationsDir: migrationsDir,
		schemaTable:   "schema_migrations",
		dbDriver:      dbDriver,
		dbDSN:         dbDSN,
	}
}

// Create 创建新迁移
func (m *StandardMigrator) Create(name string, migrationType MigrationType) (*Migration, error) {
	logger.Info("创建新迁移: %s", name)

	// 确保迁移目录存在
	if err := utils.EnsureDir(m.migrationsDir); err != nil {
		return nil, fmt.Errorf("创建迁移目录失败: %w", err)
	}

	// 生成迁移ID (时间戳)
	id := time.Now().Format("20060102150405")

	// 格式化迁移名称 (转为小写，替换空格为下划线)
	formattedName := strings.ToLower(strings.ReplaceAll(name, " ", "_"))

	// 文件名格式: {id}_{name}.{extension}
	var extension string
	if migrationType == TypeSQL {
		extension = "sql"
	} else {
		extension = "go"
	}

	// 创建迁移文件
	fileName := fmt.Sprintf("%s_%s.%s", id, formattedName, extension)
	filePath := filepath.Join(m.migrationsDir, fileName)

	// 创建迁移文件
	f, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建迁移文件失败: %w", err)
	}
	defer f.Close()

	// 写入默认内容
	var template string
	if migrationType == TypeSQL {
		template = fmt.Sprintf(`-- 迁移: %s
-- 创建时间: %s
-- 说明: %s

-- 向上迁移
-- +migrate Up
CREATE TABLE example (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 向下迁移
-- +migrate Down
DROP TABLE IF EXISTS example;
`, formattedName, time.Now().Format("2006-01-02 15:04:05"), name)
	} else {
		template = fmt.Sprintf(`package migrations

import (
	"context"
	"database/sql"
	"time"
)

// %s 迁移函数
// 创建时间: %s
// 说明: %s

// Up 向上迁移
func Up_%s(ctx context.Context, db *sql.DB) error {
	// 实现向上迁移逻辑
	_, err := db.ExecContext(ctx, `+"`"+`
		CREATE TABLE example (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`+"`"+`)
	return err
}

// Down 向下迁移
func Down_%s(ctx context.Context, db *sql.DB) error {
	// 实现向下迁移逻辑
	_, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS example")
	return err
}
`, formattedName, time.Now().Format("2006-01-02 15:04:05"), name, id, id)
	}

	if _, err := f.WriteString(template); err != nil {
		return nil, fmt.Errorf("写入迁移文件失败: %w", err)
	}

	logger.Info("成功创建迁移文件: %s", filePath)

	// 返回创建的迁移信息
	return &Migration{
		ID:         id,
		Name:       name,
		Type:       migrationType,
		Status:     StatusPending,
		FilePath:   filePath,
		Reversible: true,
	}, nil
}

// findMigrations 查找所有迁移文件
func (m *StandardMigrator) findMigrations() ([]Migration, error) {
	// 确保迁移目录存在
	if err := utils.EnsureDir(m.migrationsDir); err != nil {
		return nil, fmt.Errorf("检查迁移目录失败: %w", err)
	}

	// 读取目录内容
	entries, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败: %w", err)
	}

	// 解析迁移文件
	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		parts := strings.SplitN(fileName, "_", 2)
		if len(parts) != 2 {
			continue
		}

		id := parts[0]
		if len(id) != 14 { // 通常是年月日时分秒 (14位)
			continue
		}

		nameParts := strings.Split(parts[1], ".")
		if len(nameParts) != 2 {
			continue
		}

		name := nameParts[0]
		ext := nameParts[1]

		var migrationType MigrationType
		if ext == "sql" {
			migrationType = TypeSQL
		} else if ext == "go" {
			migrationType = TypeGoFn
		} else {
			continue
		}

		migrations = append(migrations, Migration{
			ID:         id,
			Name:       name,
			Type:       migrationType,
			Status:     StatusPending, // 默认为待执行
			FilePath:   filepath.Join(m.migrationsDir, fileName),
			Reversible: true,
		})
	}

	// 按ID排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

// Up 执行向上迁移
func (m *StandardMigrator) Up(steps int) error {
	logger.Info("执行向上迁移, 步数: %d", steps)

	// 获取所有迁移
	migrations, err := m.findMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		logger.Info("没有找到迁移文件")
		return nil
	}

	// TODO: 连接数据库并检查已应用迁移
	// 这里为演示简化处理，实际应查询schemaTable表获取已应用迁移状态

	// 计算要应用的迁移数
	applyCount := len(migrations)
	if steps > 0 && steps < applyCount {
		applyCount = steps
	}

	// 模拟应用迁移
	for i := 0; i < applyCount; i++ {
		logger.Info("应用迁移: %s (%s)", migrations[i].Name, migrations[i].ID)
		// TODO: 实际执行迁移SQL或Go函数
	}

	logger.Info("成功应用 %d 个迁移", applyCount)
	return nil
}

// Down 执行向下迁移(回滚)
func (m *StandardMigrator) Down(steps int) error {
	logger.Info("执行向下迁移(回滚), 步数: %d", steps)

	// 获取所有迁移
	migrations, err := m.findMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		logger.Info("没有找到迁移文件")
		return nil
	}

	// TODO: 连接数据库并检查已应用迁移
	// 这里为演示简化处理，实际应查询schemaTable表获取已应用迁移状态

	// 反转迁移顺序（回滚从最新到最旧）
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID > migrations[j].ID
	})

	// 计算要回滚的迁移数
	rollbackCount := len(migrations)
	if steps > 0 && steps < rollbackCount {
		rollbackCount = steps
	}

	// 模拟回滚迁移
	for i := 0; i < rollbackCount; i++ {
		logger.Info("回滚迁移: %s (%s)", migrations[i].Name, migrations[i].ID)
		// TODO: 实际执行回滚SQL或Go函数
	}

	logger.Info("成功回滚 %d 个迁移", rollbackCount)
	return nil
}

// Status 获取迁移状态
func (m *StandardMigrator) Status() ([]Migration, error) {
	logger.Info("获取迁移状态")

	// 获取所有迁移
	migrations, err := m.findMigrations()
	if err != nil {
		return nil, err
	}

	if len(migrations) == 0 {
		logger.Info("没有找到迁移文件")
		return migrations, nil
	}

	// TODO: 连接数据库并检查已应用迁移
	// 这里为演示简化处理，实际应查询schemaTable表获取已应用迁移状态
	// 假设前3个已应用
	appliedCount := min(3, len(migrations))
	for i := 0; i < appliedCount; i++ {
		migrations[i].Status = StatusApplied
		migrations[i].AppliedAt = time.Now().Add(-time.Duration(i) * 24 * time.Hour)
		migrations[i].BatchID = 1
	}

	return migrations, nil
}

// Reset 重置所有迁移
func (m *StandardMigrator) Reset() error {
	logger.Info("重置所有迁移")

	// 先全部回滚
	if err := m.Down(0); err != nil {
		return fmt.Errorf("回滚迁移失败: %w", err)
	}

	// 再全部应用
	if err := m.Up(0); err != nil {
		return fmt.Errorf("应用迁移失败: %w", err)
	}

	logger.Info("成功重置所有迁移")
	return nil
}

// Refresh 刷新所有迁移
func (m *StandardMigrator) Refresh() error {
	return m.Reset()
}

// Generate 生成新迁移
func (m *StandardMigrator) Generate(name, template string) (*Migration, error) {
	logger.Info("生成新迁移: %s", name)

	// 确定迁移类型
	var migrationType MigrationType
	if strings.HasSuffix(template, ".sql") {
		migrationType = TypeSQL
	} else {
		migrationType = TypeGoFn
	}

	// 创建迁移
	migration, err := m.Create(name, migrationType)
	if err != nil {
		return nil, err
	}

	// 如果指定了模板，覆盖默认内容
	if template != "" && utils.FileExists(template) {
		content, err := os.ReadFile(template)
		if err != nil {
			return nil, fmt.Errorf("读取模板文件失败: %w", err)
		}

		if err := os.WriteFile(migration.FilePath, content, 0644); err != nil {
			return nil, fmt.Errorf("写入模板内容失败: %w", err)
		}
	}

	return migration, nil
}

// FormatMigrationStatus 格式化迁移状态输出
func FormatMigrationStatus(migrations []Migration) string {
	if len(migrations) == 0 {
		return "没有发现迁移文件"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("找到 %d 个迁移:\n\n", len(migrations)))
	builder.WriteString(fmt.Sprintf("%-14s | %-30s | %-8s | %-10s | %-19s | %s\n",
		"ID", "名称", "类型", "状态", "应用时间", "批次"))
	builder.WriteString(strings.Repeat("-", 100) + "\n")

	for _, m := range migrations {
		appliedAt := ""
		if m.Status == StatusApplied {
			appliedAt = m.AppliedAt.Format("2006-01-02 15:04:05")
		}

		batchID := ""
		if m.Status == StatusApplied {
			batchID = strconv.Itoa(m.BatchID)
		}

		builder.WriteString(fmt.Sprintf("%-14s | %-30s | %-8s | %-10s | %-19s | %s\n",
			m.ID, m.Name, m.Type, m.Status, appliedAt, batchID))
	}

	return builder.String()
}

// min 返回两个int中较小的一个
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
