package data

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Migrate 执行数据库迁移
func Migrate(dataSource string, migrateDir string, autoRun bool) error {
	// 1. 验证配置
	if migrateDir == "" {
		return fmt.Errorf("数据库迁移目录未配置")
	}

	if !autoRun {
		logx.Info("AutoRun 已关闭，跳过数据库迁移")
		return nil
	}

	// 2. 使用 go-zero 的 sqlx 创建数据库连接
	db, err := sqlx.NewMysql(dataSource).RawDB()
	if err != nil {
		return fmt.Errorf("获取原始数据库连接失败: %w", err)
	}

	// 3. 初始化迁移驱动
	driver, err := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: "schema_migration", // 迁移记录表
	})
	if err != nil {
		return fmt.Errorf("初始化迁移驱动失败: %w", err)
	}

	// 4. 执行迁移
	source := fmt.Sprintf("file://%s", filepath.ToSlash(migrateDir))
	m, err := migrate.NewWithDatabaseInstance(source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}

	// 5. 执行升级操作
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logx.Info("数据库已处于最新版本，无需迁移")
			return nil
		}
		return fmt.Errorf("迁移执行失败: %w", err)
	}

	logx.Info("数据库迁移成功")
	return nil
}
