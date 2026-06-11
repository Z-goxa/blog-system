package main

import (
	"fmt"
	"log"
	"time"
)

// SeedData 添加示例数据
func SeedData() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 检查是否已经有数据
	var postCount int64
	db.Model(&Post{}).Count(&postCount)
	
	if postCount > 0 {
		log.Println("数据库已包含数据，跳过初始化")
		return nil
	}

	log.Println("开始初始化示例数据...")

	// 创建示例分类
	categories := []Category{
		{Name: "Wails", Slug: "wails", Description: "Wails 框架相关文章"},
		{Name: "Go", Slug: "go", Description: "Go 语言特性和开发经验"},
		{Name: "Vue", Slug: "vue", Description: "Vue 前端开发"},
		{Name: "技术杂谈", Slug: "tech", Description: "其他技术相关文章"},
	}

	for _, cat := range categories {
		var existing Category
		db.Where("slug = ?", cat.Slug).First(&existing)
		if existing.ID == 0 {
			db.Create(&cat)
			log.Printf("创建分类: %s", cat.Name)
		}
	}

	// 创建示例标签
	tags := []Tag{
		{Name: "Wails", Slug: "wails"},
		{Name: "Go", Slug: "go"},
		{Name: "Vue", Slug: "vue"},
		{Name: "桌面应用", Slug: "desktop-app"},
		{Name: "语言特性", Slug: "language-features"},
		{Name: "前端", Slug: "frontend"},
	}

	for _, tag := range tags {
		var existing Tag
		db.Where("slug = ?", tag.Slug).First(&existing)
		if existing.ID == 0 {
			db.Create(&tag)
			log.Printf("创建标签: %s", tag.Name)
		}
	}

	// 获取分类和标签的ID
	var wailsCat, goCat, vueCat, techCat Category
	db.Where("slug = ?", "wails").First(&wailsCat)
	db.Where("slug = ?", "go").First(&goCat)
	db.Where("slug = ?", "vue").First(&vueCat)
	db.Where("slug = ?", "tech").First(&techCat)

	var wailsTag, goTag, vueTag, desktopTag, langTag, frontendTag Tag
	db.Where("slug = ?", "wails").First(&wailsTag)
	db.Where("slug = ?", "go").First(&goTag)
	db.Where("slug = ?", "vue").First(&vueTag)
	db.Where("slug = ?", "desktop-app").First(&desktopTag)
	db.Where("slug = ?", "language-features").First(&langTag)
	db.Where("slug = ?", "frontend").First(&frontendTag)

	// 创建示例文章
	now := time.Now()
	posts := []Post{
		{
			UserID:          1,
			CategoryID:      &wailsCat.ID,
			Title:           "Wails 2.0 正式发布",
			Slug:            "wails-2-0-release",
			Excerpt:         "Wails 2.0 带来了全新的架构设计，提供更好的性能和开发体验。",
			ContentMarkdown: "Wails 2.0 是一个革命性的桌面应用开发框架，它允许开发者使用 Go 语言和现代前端技术构建高性能的桌面应用。",
			Status:      "published",
			PublishedAt: &now,
			Meta: JSONMap{
				"cover_image": "https://wails.io/img/wails-logo.png",
				"source_url": "https://wails.io/zh-CN/blog/wails-2-0",
			},
		},
		{
			UserID:          1,
			CategoryID:      &goCat.ID,
			Title:           "Go 1.21 新特性详解",
			Slug:            "go-1-21-features",
			Excerpt:         "Go 1.21 版本带来了许多令人兴奋的新特性，包括结构化日志、泛型改进、性能优化等。",
			ContentMarkdown: "Go 1.21 版本带来了许多令人兴奋的新特性，包括结构化日志、泛型改进、性能优化等。",
			Status:      "published",
			PublishedAt: func() *time.Time { t := now.Add(-24 * time.Hour); return &t }(),
			Meta: JSONMap{
				"cover_image": "https://go.dev/blog/go-brand/Go-Logo/SVG/Go-Logo_LightBlue.svg",
				"source_url": "https://go.dev/blog/go1.21",
			},
		},
		{
			UserID:          1,
			CategoryID:      &vueCat.ID,
			Title:           "Vue 3.4 新特性解析",
			Slug:            "vue-3-4-new-features",
			Excerpt:         "Vue 3.4 版本带来了响应式系统的改进、编译器优化和更好的开发者体验。",
			ContentMarkdown: "Vue 3.4 版本是 Vue 3 系列的重要更新，带来了许多令人兴奋的新特性和改进。",
			Status:      "published",
			PublishedAt: func() *time.Time { t := now.Add(-48 * time.Hour); return &t }(),
			Meta: JSONMap{
				"cover_image": "https://vuejs.org/logo.png",
				"source_url": "https://blog.vuejs.org/",
			},
		},
		{
			UserID:          1,
			CategoryID:      &techCat.ID,
			Title:           "使用 Wails 构建跨平台桌面应用",
			Slug:            "build-cross-platform-desktop-app",
			Excerpt:         "Wails 是一个优秀的 Go 语言桌面应用开发框架，它允许开发者使用现代 Web 技术构建桌面应用。",
			ContentMarkdown: "Wails 提供了一种现代、高效的方式来构建跨平台桌面应用。它结合了 Go 语言的性能优势和 Web 技术的灵活性。",
			Status:      "published",
			PublishedAt: func() *time.Time { t := now.Add(-72 * time.Hour); return &t }(),
			Meta: JSONMap{
				"cover_image": "https://wails.io/img/wails-logo.png",
				"source_url": "https://wails.io/zh-CN/blog/cross-platform",
			},
		},
	}

	// 创建文章并关联标签
	for _, post := range posts {
		// 渲染 HTML 内容
		post.ContentHTML = renderMarkdown(post.ContentMarkdown)
		db.Create(&post)
		log.Printf("创建文章: %s", post.Title)

		// 关联标签
		var postTags []Tag
		switch post.CategoryID {
		case &wailsCat.ID:
			postTags = []Tag{wailsTag, goTag, desktopTag}
		case &goCat.ID:
			postTags = []Tag{goTag, langTag}
		case &vueCat.ID:
			postTags = []Tag{vueTag, frontendTag}
		case &techCat.ID:
			postTags = []Tag{wailsTag, goTag, desktopTag}
		}

		for _, tag := range postTags {
			db.Exec("INSERT INTO post_tag_relation (post_id, tag_id) VALUES (?, ?)", post.ID, tag.ID)
		}
	}

	log.Println("示例数据初始化完成！")
	return nil
}
