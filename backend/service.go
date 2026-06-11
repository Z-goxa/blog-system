package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gorm.io/gorm"
)

// mdRenderer 全局 markdown 渲染器（GFM + 表格 + 代码高亮 class）
var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Table),
	goldmark.WithRendererOptions(html.WithHardWraps(), html.WithUnsafe()),
)

// renderMarkdown 将 markdown 转为 HTML（失败时返回原始文本、不阻断流程）
func renderMarkdown(md string) string {
	if md == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(md), &buf); err != nil {
		log.Printf("⚠️  markdown 渲染失败: %v", err)
		return md
	}
	return buf.String()
}

// BlogService 博客服务
type BlogService struct {
	db *gorm.DB
}

// NewBlogService 创建服务实例
func NewBlogService() *BlogService {
	return &BlogService{
		db: GetDB(),
	}
}

// CreatePost 创建文章（包含事务处理），返回新文章 ID
func (s *BlogService) CreatePost(title, slug, content string, categoryId uint, tags []string, status string) (uint, error) {
	// 参数校验
	if title == "" {
		return 0, errors.New("标题不能为空")
	}
	if content == "" {
		return 0, errors.New("内容不能为空")
	}
	if status == "" {
		status = "draft"
	}

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("开启事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建文章
	var categoryIDPtr *uint
	if categoryId > 0 {
		categoryIDPtr = &categoryId
	}

	if slug == "" {
		slug = generateSlug(title)
	}

	var publishedAt *time.Time
	if status == "published" {
		now := time.Now()
		publishedAt = &now
	}

	// 获取默认管理员用户 ID (Wails 桌面端通常只有一个活跃用户)
	var admin User
	if err := tx.Where("role = ?", "admin").First(&admin).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("未找到管理员用户，请先登录: %w", err)
	}

	post := Post{
		UserID:          admin.ID, // 设置 UserID
		Title:           title,
		ContentMarkdown: content,
		ContentHTML:     renderMarkdown(content),
		CategoryID:      categoryIDPtr,
		Status:          status,
		Slug:            slug,
		PublishedAt:     publishedAt,
	}

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("创建文章失败: %w", err)
	}

	// 2. 处理标签关联
	if len(tags) > 0 {
		var tagModels []Tag
		for _, tagName := range tags {
			var tag Tag
			if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, Tag{
				Name: tagName,
				Slug: generateSlug(tagName),
			}).Error; err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("处理标签失败: %w", err)
			}
			tagModels = append(tagModels, tag)
		}

		if err := tx.Model(&post).Association("Tags").Append(tagModels); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("关联标签失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return post.ID, nil
}

// UpdatePost 更新文章
func (s *BlogService) UpdatePost(id uint, title, slug, content string, categoryId uint, tags []string, status string) error {
	var post Post
	if err := s.db.First(&post, id).Error; err != nil {
		return err
	}

	tx := s.db.Begin()

	var categoryIDPtr *uint
	if categoryId > 0 {
		categoryIDPtr = &categoryId
	}

	updates := map[string]interface{}{
		"Title":           title,
		"Slug":            slug,
		"ContentMarkdown": content,
		"ContentHTML":     renderMarkdown(content),
		"CategoryID":      categoryIDPtr,
		"Status":          status,
	}

	// 如果是从草稿变更为发布，设置发布时间
	if post.Status != "published" && status == "published" {
		now := time.Now()
		updates["PublishedAt"] = &now
	}

	if err := tx.Model(&post).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 处理标签
	if err := tx.Model(&post).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	if len(tags) > 0 {
		var tagModels []Tag
		for _, tagName := range tags {
			var tag Tag
			tx.Where("name = ?", tagName).FirstOrCreate(&tag, Tag{Name: tagName, Slug: generateSlug(tagName)})
			tagModels = append(tagModels, tag)
		}
		tx.Model(&post).Association("Tags").Append(tagModels)
	}

	return tx.Commit().Error
}

// PostDTO 文章数据传输对象
type PostDTO struct {
	ID          uint         `json:"id"`
	Title       string       `json:"title"`
	Slug        string       `json:"slug"`
	Excerpt     string       `json:"excerpt"`
	ContentHTML string       `json:"content_html"`
	Status      string       `json:"status"`
	ViewCount   uint         `json:"view_count"`
	Category    *CategoryDTO `json:"category,omitempty"`
	Tags        []TagDTO     `json:"tags,omitempty"`
	PublishedAt *time.Time   `json:"published_at"`
	CreatedAt   time.Time    `json:"created_at"`
}

type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PublicCategoryDTO struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"post_count"`
}

type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// GetPublishedPosts 获取已发布文章（带分页）
func (s *BlogService) GetPublishedPosts(page, pageSize int) ([]PostDTO, int64, error) {
	return s.GetPublishedPostsByCategory(page, pageSize, "", "", "")
}

// GetPublishedPostsByCategory 获取已发布文章，可按分类 slug 和提示词/关键词过滤
func (s *BlogService) GetPublishedPostsByCategory(page, pageSize int, categorySlug string, search string, archive string) ([]PostDTO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var posts []Post
	var total int64

	countQuery := s.db.Model(&Post{}).Where("posts.status = ?", "published")
	listQuery := s.db.Where("posts.status = ?", "published")
	if categorySlug != "" {
		countQuery = countQuery.Joins("JOIN categories ON categories.id = posts.category_id").Where("categories.slug = ?", categorySlug)
		listQuery = listQuery.Joins("JOIN categories ON categories.id = posts.category_id").Where("categories.slug = ?", categorySlug)
	}
	if search != "" {
		like := "%" + search + "%"
		condition := "(posts.title LIKE ? OR posts.excerpt LIKE ? OR posts.content_markdown LIKE ?)"
		countQuery = countQuery.Where(condition, like, like, like)
		listQuery = listQuery.Where(condition, like, like, like)
	}
	if archive != "" {
		// archive format: YYYY-MM
		t, err := time.Parse("2006-01", archive)
		if err == nil {
			y, m := t.Year(), t.Month()
			nextM := m + 1
			nextY := y
			if nextM > 12 {
				nextM = 1
				nextY++
			}
			startDate := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
			endDate := time.Date(nextY, nextM, 1, 0, 0, 0, 0, time.UTC)
			cond := "COALESCE(posts.published_at, posts.created_at) >= ? AND COALESCE(posts.published_at, posts.created_at) < ?"
			countQuery = countQuery.Where(cond, startDate, endDate)
			listQuery = listQuery.Where(cond, startDate, endDate)
		}
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询文章总数失败: %w", err)
	}

	if err := listQuery.
		Preload("Category").
		Preload("Tags").
		Preload("User").
		Order("posts.published_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("查询文章列表失败: %w", err)
	}

	postDTOs := []PostDTO{}
	for _, post := range posts {
		excerpt := post.Excerpt
		if excerpt == "" {
			excerpt = makeExcerpt(post.ContentMarkdown, 180)
		}
		dto := PostDTO{
			ID:          post.ID,
			Title:       post.Title,
			Slug:        post.Slug,
			Excerpt:     excerpt,
			Status:      post.Status,
			ViewCount:   post.ViewCount,
			PublishedAt: post.PublishedAt,
			CreatedAt:   post.CreatedAt,
		}
		if post.Category != nil {
			dto.Category = &CategoryDTO{
				ID:   post.Category.ID,
				Name: post.Category.Name,
				Slug: post.Category.Slug,
			}
		}
		for _, tag := range post.Tags {
			dto.Tags = append(dto.Tags, TagDTO{
				ID:   tag.ID,
				Name: tag.Name,
				Slug: tag.Slug,
			})
		}
		postDTOs = append(postDTOs, dto)
	}

	return postDTOs, total, nil
}

// GetPublishedArchives 获取已发布文章归档月份
func (s *BlogService) GetPublishedArchives() ([]ArchiveDTO, error) {
	type row struct {
		ArchiveKey string
		Count      int64
	}
	var rows []row
	if err := s.db.Model(&Post{}).
		Select("DATE_FORMAT(COALESCE(published_at, created_at), '%Y-%m') AS archive_key, COUNT(*) AS count").
		Where("status = ? AND COALESCE(published_at, created_at) IS NOT NULL", "published").
		Group("archive_key").
		Order("archive_key DESC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("查询文章归档失败: %w", err)
	}

	archives := make([]ArchiveDTO, 0, len(rows))
	for _, r := range rows {
		if r.ArchiveKey == "" {
			continue
		}
		t, err := time.Parse("2006-01", r.ArchiveKey)
		label := r.ArchiveKey
		if err == nil {
			label = fmt.Sprintf("%d年%d月", t.Year(), int(t.Month()))
		}
		archives = append(archives, ArchiveDTO{Key: r.ArchiveKey, Label: label, Count: r.Count})
	}
	return archives, nil
}

// GetPostBySlug 根据 Slug 获取文章详情
func (s *BlogService) GetPostBySlug(slug string) (*PostDTO, error) {
	if slug == "" {
		return nil, errors.New("slug 不能为空")
	}

	var post Post
	if err := s.db.
		Where("slug = ? AND status = ?", slug, "published").
		Preload("Category").
		Preload("Tags").
		Preload("User").
		First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, fmt.Errorf("查询文章失败: %w", err)
	}

	s.db.Model(&post).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	excerpt := post.Excerpt
	if excerpt == "" {
		excerpt = makeExcerpt(post.ContentMarkdown, 180)
	}

	dto := &PostDTO{
		ID:          post.ID,
		Title:       post.Title,
		Slug:        post.Slug,
		Excerpt:     excerpt,
		ContentHTML: post.ContentHTML,
		Status:      post.Status,
		ViewCount:   post.ViewCount + 1,
		PublishedAt: post.PublishedAt,
		CreatedAt:   post.CreatedAt,
	}

	if post.Category != nil {
		dto.Category = &CategoryDTO{
			ID:   post.Category.ID,
			Name: post.Category.Name,
			Slug: post.Category.Slug,
		}
	}
	for _, tag := range post.Tags {
		dto.Tags = append(dto.Tags, TagDTO{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	return dto, nil
}

// GetTags 获取所有标签
func (s *BlogService) GetTags() ([]Tag, error) {
	var tags []Tag
	if err := s.db.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}
	return tags, nil
}

// CreateTag 创建标签
func (s *BlogService) CreateTag(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("标签名称不能为空")
	}
	var existing Tag
	if err := s.db.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查标签失败: %w", err)
	}
	slug := strings.TrimSpace(generateStableSlug(name))
	if slug == "" {
		slug = fmt.Sprintf("tag-%x", []byte(name))
		if len(slug) > 50 {
			slug = slug[:50]
		}
	}
	baseSlug := slug
	for i := 2; ; i++ {
		var count int64
		if err := s.db.Model(&Tag{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return fmt.Errorf("检查标签 slug 失败: %w", err)
		}
		if count == 0 {
			break
		}
		suffix := fmt.Sprintf("-%d", i)
		maxBaseLen := 50 - len(suffix)
		if maxBaseLen < 1 {
			maxBaseLen = 1
		}
		truncated := baseSlug
		if len([]rune(truncated)) > maxBaseLen {
			truncated = string([]rune(truncated)[:maxBaseLen])
		}
		slug = truncated + suffix
	}
	tag := Tag{Name: name, Slug: slug}
	return s.db.Create(&tag).Error
}

// DeleteTag 删除标签
func (s *BlogService) DeleteTag(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除文章-标签关联，避免 MariaDB 外键约束阻止删除标签。
		if err := tx.Exec("DELETE FROM post_tag_relation WHERE tag_id = ?", id).Error; err != nil {
			return fmt.Errorf("删除标签关联失败: %w", err)
		}
		if err := tx.Delete(&Tag{}, id).Error; err != nil {
			return fmt.Errorf("删除标签失败: %w", err)
		}
		return nil
	})
}

// GetCategories 获取所有分类
func (s *BlogService) GetCategories() ([]Category, error) {
	var categories []Category
	if err := s.db.Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	return categories, nil
}

// GetPublishedCategories 获取公开分类导航：返回数据库全部分类，并附带已发布文章数量
func (s *BlogService) GetPublishedCategories() ([]PublicCategoryDTO, error) {
	var categories []PublicCategoryDTO
	if err := s.db.Table("categories").
		Select("categories.id, categories.name, categories.slug, COUNT(posts.id) AS post_count").
		Joins("LEFT JOIN posts ON posts.category_id = categories.id AND posts.status = ?", "published").
		Group("categories.id, categories.name, categories.slug, categories.sort_order").
		Order("categories.sort_order ASC, categories.id ASC").
		Scan(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询公开分类失败: %w", err)
	}
	return categories, nil
}

// CreateCategory 创建分类
func (s *BlogService) CreateCategory(name, slug, description string) error {
	if name == "" {
		return errors.New("分类名称不能为空")
	}
	if slug == "" {
		slug = generateSlug(name)
	}
	category := Category{
		Name:        name,
		Slug:        slug,
		Description: description,
	}
	return s.db.Create(&category).Error
}

// UpdateCategory 更新分类
func (s *BlogService) UpdateCategory(id uint, name, slug, description string) error {
	var category Category
	if err := s.db.First(&category, id).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{
		"Name":        name,
		"Slug":        slug,
		"Description": description,
	}
	return s.db.Model(&category).Updates(updates).Error
}

// DeleteCategory 删除分类
func (s *BlogService) DeleteCategory(id uint) error {
	// 检查是否有文章使用此分类
	var count int64
	s.db.Model(&Post{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该分类下已有文章，无法删除")
	}
	return s.db.Delete(&Category{}, id).Error
}

// GetPostByID 根据 ID 获取文章详情
func (s *BlogService) GetPostByID(id uint) (*Post, error) {
	var post Post
	if err := s.db.Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// GetAllPosts 获取所有文章（管理后台用）
func (s *BlogService) GetAllPosts() ([]PostDTO, error) {
	var posts []Post
	if err := s.db.Preload("Category").Preload("Tags").Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("查询文章列表失败: %w", err)
	}

	var postDTOs []PostDTO
	for _, post := range posts {
		dto := PostDTO{
			ID:          post.ID,
			Title:       post.Title,
			Slug:        post.Slug,
			Status:      post.Status,
			ViewCount:   post.ViewCount,
			PublishedAt: post.PublishedAt,
			CreatedAt:   post.CreatedAt,
		}
		if post.Category != nil {
			dto.Category = &CategoryDTO{
				ID:   post.Category.ID,
				Name: post.Category.Name,
				Slug: post.Category.Slug,
			}
		}
		postDTOs = append(postDTOs, dto)
	}
	return postDTOs, nil
}

// DeletePost 删除文章
func (s *BlogService) DeletePost(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 先删除文章-标签关联，避免 MariaDB 外键约束阻止删除文章。
		if err := tx.Table("post_tag_relation").Where("post_id = ?", id).Delete(nil).Error; err != nil {
			return fmt.Errorf("删除文章标签关联失败: %w", err)
		}
		if err := tx.Delete(&Post{}, id).Error; err != nil {
			return fmt.Errorf("删除文章失败: %w", err)
		}
		return nil
	})
}

// CommentDTO 评论数据传输对象
type CommentDTO struct {
	ID          uint      `json:"id"`
	PostTitle   string    `json:"post_title"`
	PostSlug    string    `json:"post_slug"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetCommentsByPostSlug 获取文章已通过评论
func (s *BlogService) GetCommentsByPostSlug(slug string) ([]CommentDTO, error) {
	var post Post
	if err := s.db.Select("id").Where("slug = ? AND status = ?", slug, "published").First(&post).Error; err != nil {
		return []CommentDTO{}, fmt.Errorf("文章不存在")
	}
	var comments []Comment
	if err := s.db.Where("post_id = ? AND status = ?", post.ID, "approved").Order("created_at ASC").Find(&comments).Error; err != nil {
		return []CommentDTO{}, fmt.Errorf("查询评论失败: %w", err)
	}
	dtos := []CommentDTO{}
	for _, c := range comments {
		dtos = append(dtos, CommentDTO{
			ID:          c.ID,
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			Content:     c.Content,
			Status:      c.Status,
			CreatedAt:   c.CreatedAt,
		})
	}
	return dtos, nil
}

// CreateComment 创建前台访客评论
func (s *BlogService) CreateComment(slug string, userID uint, content string) (*CommentDTO, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, errors.New("评论内容不能超过 2000 字")
	}
	var post Post
	if err := s.db.Select("id").Where("slug = ? AND status = ?", slug, "published").First(&post).Error; err != nil {
		return nil, errors.New("文章不存在")
	}
	var user User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	authorName := user.DisplayName
	if authorName == "" {
		authorName = user.Username
	}
	comment := Comment{
		PostID:      post.ID,
		UserID:      &user.ID,
		AuthorName:  authorName,
		AuthorEmail: user.Email,
		Content:     content,
		Status:      "approved",
	}
	if err := s.db.Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("发表评论失败: %w", err)
	}
	return &CommentDTO{
		ID:          comment.ID,
		AuthorName:  comment.AuthorName,
		AuthorEmail: comment.AuthorEmail,
		Content:     comment.Content,
		Status:      comment.Status,
		CreatedAt:   comment.CreatedAt,
	}, nil
}

// GetLatestApprovedComments 获取前台最新评论
func (s *BlogService) GetLatestApprovedComments(limit int) ([]CommentDTO, error) {
	if limit < 1 || limit > 20 {
		limit = 5
	}
	dtos := []CommentDTO{}
	var comments []Comment
	if err := s.db.Where("status = ?", "approved").Order("created_at DESC").Limit(limit).Find(&comments).Error; err != nil {
		return dtos, fmt.Errorf("查询最新评论失败: %w", err)
	}

	for _, c := range comments {
		var post Post
		s.db.Select("title", "slug").Where("id = ? AND status = ?", c.PostID, "published").First(&post)
		if post.Slug == "" {
			continue
		}
		dtos = append(dtos, CommentDTO{
			ID:          c.ID,
			PostTitle:   post.Title,
			PostSlug:    post.Slug,
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			Content:     c.Content,
			Status:      c.Status,
			CreatedAt:   c.CreatedAt,
		})
	}
	return dtos, nil
}

// GetAllComments 获取所有评论
func (s *BlogService) GetAllComments() ([]CommentDTO, error) {
	dtos := []CommentDTO{} // 始终返回非 nil 切片，以免前端在 null 上访问 .length
	var comments []Comment
	if err := s.db.Order("created_at DESC").Find(&comments).Error; err != nil {
		return dtos, fmt.Errorf("查询评论失败: %w", err)
	}

	for _, c := range comments {
		var postTitle string
		s.db.Model(&Post{}).Select("title").Where("id = ?", c.PostID).Scan(&postTitle)

		dtos = append(dtos, CommentDTO{
			ID:          c.ID,
			PostTitle:   postTitle,
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			Content:     c.Content,
			Status:      c.Status,
			CreatedAt:   c.CreatedAt,
		})
	}
	return dtos, nil
}

// UpdateCommentStatus 更新评论状态
func (s *BlogService) UpdateCommentStatus(id uint, status string) error {
	return s.db.Model(&Comment{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteComment 删除评论
func (s *BlogService) DeleteComment(id uint) error {
	return s.db.Delete(&Comment{}, id).Error
}

// Stats 统计信息
type Stats struct {
	PostCount     int64 `json:"post_count"`
	CategoryCount int64 `json:"category_count"`
	TagCount      int64 `json:"tag_count"`
	CommentCount  int64 `json:"comment_count"`
}

// GetStats 获取统计信息
func (s *BlogService) GetStats() (*Stats, error) {
	var stats Stats
	if s.db == nil {
		return nil, errors.New("数据库连接未初始化")
	}
	s.db.Model(&Post{}).Count(&stats.PostCount)
	s.db.Model(&Category{}).Count(&stats.CategoryCount)
	s.db.Model(&Tag{}).Count(&stats.TagCount)
	s.db.Model(&Comment{}).Count(&stats.CommentCount)
	return &stats, nil
}

// 辅助函数：生成 Slug
func makeExcerpt(content string, limit int) string {
	text := strings.TrimSpace(content)
	// 简单去掉常见 Markdown 语法，避免首页摘要为空。
	replacements := []string{"#", "*", "`", ">", "[", "]", "(", ")", "_"}
	for _, r := range replacements {
		text = strings.ReplaceAll(text, r, "")
	}
	text = strings.Join(strings.Fields(text), " ")
	if limit > 0 && len([]rune(text)) > limit {
		r := []rune(text)
		return string(r[:limit]) + "..."
	}
	return text
}

func generateStableSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))
	reg := regexp.MustCompile(`[^a-z0-9\u4e00-\u9fa5]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return ""
	}
	runes := []rune(slug)
	if len(runes) > 50 {
		slug = string(runes[:50])
	}
	return slug
}

func generateSlug(title string) string {
	// 1. 转小写
	slug := strings.ToLower(title)
	// 2. 移除非字母数字字符（保留连字符）
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// 3. 去除首尾连字符
	slug = strings.Trim(slug, "-")
	// 4. 确保生成的 slug 是唯一的
	baseSlug := slug
	if baseSlug == "" || baseSlug == "post" {
		baseSlug = fmt.Sprintf("tag-%d", time.Now().UnixNano()/1000000)
	}
	// 5. 使用纳秒级时间戳确保唯一性
	uniqueSlug := fmt.Sprintf("%s-%d", baseSlug, time.Now().UnixNano())
	return uniqueSlug
}
