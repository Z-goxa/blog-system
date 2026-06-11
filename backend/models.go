package main

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSON 字段类型（用于 MariaDB 的 JSON 列）
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// User 用户模型
type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	Username     string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Salt         string     `gorm:"type:varchar(64)" json:"-"`
	Role         string     `gorm:"type:enum('admin','editor','author','subscriber');default:'author'" json:"role"`
	DisplayName  string     `gorm:"type:varchar(100)" json:"display_name"`
	AvatarURL    string     `gorm:"type:varchar(500)" json:"avatar_url"`
	Bio          string     `gorm:"type:text" json:"bio"`
	Status       string     `gorm:"type:enum('active','inactive','banned');default:'active'" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// Category 分类模型
type Category struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *uint     `gorm:"type:int unsigned;index" json:"parent_id"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	Meta        JSONMap   `gorm:"type:json" json:"meta"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Tag 标签模型
type Tag struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	Meta      JSONMap   `gorm:"type:json" json:"meta"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Post 文章模型
type Post struct {
	ID              uint       `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	UserID          uint       `gorm:"type:int unsigned;not null;index" json:"user_id"`
	CategoryID      *uint      `gorm:"type:int unsigned;index" json:"category_id"`
	Title           string     `gorm:"type:varchar(255);not null" json:"title"`
	Slug            string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Excerpt         string     `gorm:"type:text" json:"excerpt"`
	ContentMarkdown string     `gorm:"type:mediumtext;not null" json:"content_markdown"`
	ContentHTML     string     `gorm:"type:mediumtext" json:"content_html"`
	Status          string     `gorm:"type:enum('draft','pending','published','private','trash');default:'draft';index" json:"status"`
	Type            string     `gorm:"type:enum('post','page');default:'post'" json:"type"`
	CommentStatus   string     `gorm:"type:enum('open','closed','moderated');default:'open'" json:"comment_status"`
	ViewCount       uint       `gorm:"default:0" json:"view_count"`
	Meta            JSONMap    `gorm:"type:json" json:"meta"`
	PublishedAt     *time.Time `json:"published_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联（预加载用）
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag     `gorm:"many2many:post_tag_relation;" json:"tags,omitempty"`
}

// ArchiveDTO 归档月份
type ArchiveDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// PostTagRelation 文章标签关联
type PostTagRelation struct {
	PostID    uint      `gorm:"primaryKey;type:int unsigned" json:"post_id"`
	TagID     uint      `gorm:"primaryKey;type:int unsigned" json:"tag_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Comment 评论模型
type Comment struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;type:int unsigned" json:"id"`
	PostID      uint      `gorm:"type:int unsigned;not null;index" json:"post_id"`
	UserID      *uint     `gorm:"type:int unsigned;index" json:"user_id"`
	ParentID    *uint     `gorm:"type:int unsigned;index" json:"parent_id"`
	AuthorName  string    `gorm:"type:varchar(100)" json:"author_name"`
	AuthorEmail string    `gorm:"type:varchar(255)" json:"author_email"`
	AuthorURL   string    `gorm:"type:varchar(500)" json:"author_url"`
	IPAddress   []byte    `gorm:"type:varbinary(16)" json:"-"`
	UserAgent   string    `gorm:"type:text" json:"-"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Status      string    `gorm:"type:enum('pending','approved','spam','trash');default:'pending';index" json:"status"`
	Meta        JSONMap   `gorm:"type:json" json:"meta"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
