package main

import (
	"embed"
	"flag"
	"log"

	wails "github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	apiOnly := flag.Bool("api", false, "仅启动 HTTP API 服务（供浏览器开发模式使用）")
	flag.Parse()

	// 1. 初始化数据库
	db, err := InitDB(nil)
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	SetDB(db)

	// 2. 自动迁移
	// MariaDB 不允许 MODIFY 被外键引用的列（错误 1832）。
	// 策略：迁移前 drop 所有外键，迁移后按 schema.sql 重建。
	type fkSpec struct{ table, name, ddl string }
	knownFKs := []fkSpec{
		{"categories", "fk_categories_parent", "ALTER TABLE `categories` ADD CONSTRAINT `fk_categories_parent` FOREIGN KEY (`parent_id`) REFERENCES `categories`(`id`) ON DELETE SET NULL ON UPDATE CASCADE"},
		{"posts", "fk_posts_user", "ALTER TABLE `posts` ADD CONSTRAINT `fk_posts_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE"},
		{"posts", "fk_posts_category", "ALTER TABLE `posts` ADD CONSTRAINT `fk_posts_category` FOREIGN KEY (`category_id`) REFERENCES `categories`(`id`) ON DELETE SET NULL ON UPDATE CASCADE"},
		{"post_tag_relation", "fk_ptr_post", "ALTER TABLE `post_tag_relation` ADD CONSTRAINT `fk_ptr_post` FOREIGN KEY (`post_id`) REFERENCES `posts`(`id`) ON DELETE CASCADE ON UPDATE CASCADE"},
		{"post_tag_relation", "fk_ptr_tag", "ALTER TABLE `post_tag_relation` ADD CONSTRAINT `fk_ptr_tag` FOREIGN KEY (`tag_id`) REFERENCES `tags`(`id`) ON DELETE CASCADE ON UPDATE CASCADE"},
		{"comments", "fk_comments_post", "ALTER TABLE `comments` ADD CONSTRAINT `fk_comments_post` FOREIGN KEY (`post_id`) REFERENCES `posts`(`id`) ON DELETE CASCADE ON UPDATE CASCADE"},
		{"comments", "fk_comments_user", "ALTER TABLE `comments` ADD CONSTRAINT `fk_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE SET NULL ON UPDATE CASCADE"},
		{"comments", "fk_comments_parent", "ALTER TABLE `comments` ADD CONSTRAINT `fk_comments_parent` FOREIGN KEY (`parent_id`) REFERENCES `comments`(`id`) ON DELETE CASCADE ON UPDATE CASCADE"},
		{"sessions", "fk_sessions_user", "ALTER TABLE `sessions` ADD CONSTRAINT `fk_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE ON UPDATE CASCADE"},
	}
	hasFK := func(table, name string) bool {
		var cnt int64
		db.Raw("SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE CONSTRAINT_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ?", table, name).Scan(&cnt)
		return cnt > 0
	}
	tableExists := func(table string) bool {
		var cnt int64
		db.Raw("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", table).Scan(&cnt)
		return cnt > 0
	}
	for _, fk := range knownFKs {
		if tableExists(fk.table) && hasFK(fk.table, fk.name) {
			if err := db.Exec("ALTER TABLE `" + fk.table + "` DROP FOREIGN KEY `" + fk.name + "`").Error; err != nil {
				log.Printf("⚠️  drop 外键 %s.%s 失败: %v", fk.table, fk.name, err)
			}
		}
	}

	if err := db.AutoMigrate(
		&User{},
		&Category{},
		&Tag{},
		&Post{},
		&PostTagRelation{},
		&Comment{},
	); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}

	// 重建外键
	for _, fk := range knownFKs {
		if !tableExists(fk.table) {
			continue
		}
		if !hasFK(fk.table, fk.name) {
			if err := db.Exec(fk.ddl).Error; err != nil {
				log.Printf("⚠️  重建外键 %s.%s 失败: %v", fk.table, fk.name, err)
			}
		}
	}

	log.Println("✅ 数据库迁移完成")

	// 2.5 检查并创建默认管理员
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		passwordHash, _ := hashPassword("admin123")
		admin := User{
			Username:     "admin",
			PasswordHash: passwordHash,
			Email:        "admin@example.com",
			Role:         "admin",
			DisplayName:  "Administrator",
			Status:       "active",
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("⚠️  创建默认管理员失败: %v", err)
		} else {
			log.Println("👤 已创建默认管理员: admin / admin123")
		}
	}

	// 2.6 初始化示例种子数据（仅首次空库时）
	if err := SeedData(); err != nil {
		log.Printf("⚠️  初始化种子数据失败: %v", err)
	}

	// 2.7 启动 HTTP API 服务
	StartAPIServer()

	if *apiOnly {
		log.Println("📡 API-only 模式，按 Ctrl+C 退出")
		select {}
	}

	// 3. 运行 Wails 应用
	log.Println("🎨 启动 Wails 桌面应用")
	err = wails.Run(&options.App{
		Title:  "Blog System",
		Width:  1400,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: &UploadHandler{UploadDir: "uploads"},
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1},
		Bind: []interface{}{
			NewBlogService(),
			NewAuthService(),
			NewUploadService(), // 添加上传服务
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
