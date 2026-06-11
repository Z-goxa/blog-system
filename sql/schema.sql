-- 博客系统数据库 DDL (MariaDB 10.6+)
-- 字符集: utf8mb4 (支持 emoji 和特殊字符)
-- 存储引擎: InnoDB (支持事务和外键)

-- ============================================================
-- 1. 用户表 (users)
-- ============================================================
CREATE TABLE `users` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(50) NOT NULL COMMENT '用户名（唯一）',
    `email` VARCHAR(255) NOT NULL COMMENT '邮箱',
    `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希（Argon2id 或 bcrypt）',
    `salt` VARCHAR(64) NULL COMMENT '盐值（预留，现代算法内置盐值）',
    `role` ENUM('admin', 'editor', 'author', 'subscriber') NOT NULL DEFAULT 'author' COMMENT '角色',
    `display_name` VARCHAR(100) NULL COMMENT '显示名称',
    `avatar_url` VARCHAR(500) NULL COMMENT '头像URL',
    `bio` TEXT NULL COMMENT '个人简介',
    `status` ENUM('active', 'inactive', 'banned') NOT NULL DEFAULT 'active' COMMENT '账号状态',
    `last_login_at` DATETIME NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    UNIQUE KEY `uk_email` (`email`),
    KEY `idx_role` (`role`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='用户表'
  ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 2. 分类表 (categories)
-- ============================================================
CREATE TABLE `categories` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(100) NOT NULL COMMENT '分类名称',
    `slug` VARCHAR(100) NOT NULL COMMENT 'URL友好标识',
    `description` TEXT NULL COMMENT '分类描述',
    `parent_id` INT UNSIGNED NULL COMMENT '父分类ID（支持多级分类）',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    `meta` JSON NULL COMMENT '扩展元数据（SEO标题、描述等）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_slug` (`slug`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_sort_order` (`sort_order`),
    CONSTRAINT `fk_categories_parent` FOREIGN KEY (`parent_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='分类表';

-- ============================================================
-- 3. 标签表 (tags)
-- ============================================================
CREATE TABLE `tags` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL COMMENT '标签名称',
    `slug` VARCHAR(50) NOT NULL COMMENT 'URL友好标识',
    `description` TEXT NULL COMMENT '标签描述',
    `meta` JSON NULL COMMENT '扩展元数据',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_slug` (`slug`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='标签表';

-- ============================================================
-- 4. 文章表 (posts)
-- ============================================================
CREATE TABLE `posts` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` INT UNSIGNED NOT NULL COMMENT '作者ID',
    `category_id` INT UNSIGNED NULL COMMENT '所属分类ID',
    `title` VARCHAR(255) NOT NULL COMMENT '文章标题',
    `slug` VARCHAR(255) NOT NULL COMMENT 'URL友好标识（唯一）',
    `excerpt` TEXT NULL COMMENT '文章摘要（自动生成或手动填写）',
    `content_markdown` MEDIUMTEXT NOT NULL COMMENT 'Markdown正文',
    `content_html` MEDIUMTEXT NULL COMMENT '渲染后的HTML（缓存用）',
    `status` ENUM('draft', 'pending', 'published', 'private', 'trash') NOT NULL DEFAULT 'draft' COMMENT '文章状态',
    `type` ENUM('post', 'page') NOT NULL DEFAULT 'post' COMMENT '类型：文章或页面',
    `comment_status` ENUM('open', 'closed', 'moderated') NOT NULL DEFAULT 'open' COMMENT '评论开关',
    `view_count` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '浏览量',
    `meta` JSON NULL COMMENT '扩展元数据（SEO、自定义字段等）',
    `published_at` DATETIME NULL COMMENT '发布时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_slug` (`slug`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_category_id` (`category_id`),
    KEY `idx_status_published` (`status`, `published_at`) COMMENT '复合索引：用于查询已发布文章并按时间排序',
    KEY `idx_type_status` (`type`, `status`, `published_at`),
    KEY `idx_view_count` (`view_count`),
    FULLTEXT INDEX `ft_title_content` (`title`, `content_markdown`) COMMENT '全文搜索索引（MariaDB支持）',
    CONSTRAINT `fk_posts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_posts_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='文章表'
  ROW_FORMAT=DYNAMIC;

-- ============================================================
-- 5. 文章标签关联表 (post_tag_relation)
-- ============================================================
CREATE TABLE `post_tag_relation` (
    `post_id` INT UNSIGNED NOT NULL COMMENT '文章ID',
    `tag_id` INT UNSIGNED NOT NULL COMMENT '标签ID',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`post_id`, `tag_id`),
    KEY `idx_tag_id` (`tag_id`),
    CONSTRAINT `fk_ptr_post` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_ptr_tag` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='文章标签多对多关联表';

-- ============================================================
-- 6. 评论表 (comments)
-- ============================================================
CREATE TABLE `comments` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `post_id` INT UNSIGNED NOT NULL COMMENT '文章ID',
    `user_id` INT UNSIGNED NULL COMMENT '评论用户ID（已登录用户）',
    `parent_id` INT UNSIGNED NULL COMMENT '父评论ID（支持两级嵌套）',
    `author_name` VARCHAR(100) NULL COMMENT '访客姓名',
    `author_email` VARCHAR(255) NULL COMMENT '访客邮箱',
    `author_url` VARCHAR(500) NULL COMMENT '访客网站',
    `ip_address` VARBINARY(16) NULL COMMENT 'IPv4/IPv6地址（VARBINARY存储）',
    `user_agent` TEXT NULL COMMENT '浏览器UserAgent',
    `content` TEXT NOT NULL COMMENT '评论内容',
    `status` ENUM('pending', 'approved', 'spam', 'trash') NOT NULL DEFAULT 'pending' COMMENT '审核状态',
    `meta` JSON NULL COMMENT '扩展元数据（点赞数、举报数等）',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`id`),
    KEY `idx_post_id` (`post_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_parent_id` (`parent_id`),
    KEY `idx_status` (`status`),
    KEY `idx_post_status` (`post_id`, `status`, `created_at`) COMMENT '复合索引：查询某文章的评论（按状态和时间）',
    KEY `idx_created_at` (`created_at`),
    CONSTRAINT `fk_comments_post` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_comments_parent` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='评论表';

-- ============================================================
-- 可选：会话表 (sessions) - 用于记住登录状态
-- ============================================================
CREATE TABLE `sessions` (
    `id` VARCHAR(128) NOT NULL COMMENT '会话ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `ip_address` VARBINARY(16) NULL COMMENT '登录IP',
    `user_agent` TEXT NULL COMMENT '浏览器标识',
    `payload` JSON NOT NULL COMMENT '会话数据',
    `last_activity` INT UNSIGNED NOT NULL COMMENT '最后活动时间戳',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_last_activity` (`last_activity`),
    CONSTRAINT `fk_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci 
  COMMENT='用户会话表';

-- ============================================================
-- 初始化数据
-- ============================================================

-- 创建管理员用户（密码：Admin@123，需要用 Argon2id 哈希）
-- 实际使用时请通过程序生成哈希值
INSERT INTO `users` (`username`, `email`, `password_hash`, `role`, `display_name`, `status`) VALUES
('admin', 'admin@example.com', '$argon2id$v=19$m=65536,t=4,p=3$...', 'admin', '管理员', 'active');

-- 创建示例分类
INSERT INTO `categories` (`name`, `slug`, `description`) VALUES
('技术', 'tech', '技术相关文章'),
('生活', 'life', '生活随笔')；

-- 创建示例标签
INSERT INTO `tags` (`name`, `slug`) VALUES
('Go', 'go'),
('Wails', 'wails'),
('Vue3', 'vue3'),
('MariaDB', 'mariadb')；
