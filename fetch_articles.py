#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
文章抓取脚本（MariaDB 版）

功能：
- 从网上抓取文章链接和正文
- 直接保存到当前 blog-system 使用的 MariaDB 数据库
- 自动创建分类、标签、文章和文章-标签关联
- 读取项目根目录 .env 中的 DB_* 配置

注意：
- 默认保存为 published
- 为避免版权风险，脚本默认会保存抓取到的正文；如果用于公开博客，建议优先抓取官方/允许转载来源，或改造为摘要/整理入库。
"""

import argparse
import hashlib
import json
import os
import random
import re
import sys
import time
from datetime import datetime
from pathlib import Path
from urllib.parse import urljoin, urlparse

import pymysql
import requests
from bs4 import BeautifulSoup

PROJECT_ROOT = Path(__file__).resolve().parent
ENV_PATH = PROJECT_ROOT / ".env"

USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36",
    "Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0",
]

ARTICLE_SOURCES = [
    {
        "name": "Go 官方博客",
        "url": "https://go.dev/blog",
        "category": "Go",
        "tags": ["Go", "语言特性"],
        "include_keywords": ["/blog/"],
    },
    {
        "name": "Vue 官方博客",
        "url": "https://blog.vuejs.org/",
        "category": "Vue",
        "tags": ["Vue", "前端"],
        "include_keywords": ["blog.vuejs.org/posts/"],
    },
]


def load_env(path: Path) -> dict:
    env = {}
    if not path.exists():
        return env
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        env[key.strip()] = value.strip().strip('"').strip("'")
    return env


def get_db_config() -> dict:
    env = load_env(ENV_PATH)
    return {
        "host": env.get("DB_HOST", "localhost"),
        "port": int(env.get("DB_PORT", "3306")),
        "user": env.get("DB_USER", "root"),
        "password": env.get("DB_PASSWORD", ""),
        "database": env.get("DB_NAME", "blog_db"),
        "charset": "utf8mb4",
        "autocommit": False,
        "cursorclass": pymysql.cursors.DictCursor,
    }


def get_conn():
    return pymysql.connect(**get_db_config())


def get_random_headers():
    return {
        "User-Agent": random.choice(USER_AGENTS),
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.6",
        "Referer": "https://www.google.com/",
    }


def delay(seconds=1.5):
    time.sleep(random.uniform(seconds, seconds + 0.8))


def slugify(text: str, fallback_prefix="post") -> str:
    text = (text or "").strip().lower()
    # 中文标题也允许保留字母数字和连字符；纯中文会得到空 slug，使用 hash 兜底
    slug = re.sub(r"[^a-z0-9\u4e00-\u9fa5]+", "-", text).strip("-")
    if not slug:
        slug = f"{fallback_prefix}-{hashlib.md5(text.encode('utf-8')).hexdigest()[:10]}"
    return slug[:180]


def unique_slug(conn, table: str, base_slug: str) -> str:
    slug = base_slug
    with conn.cursor() as cur:
        i = 2
        while True:
            cur.execute(f"SELECT id FROM `{table}` WHERE slug=%s LIMIT 1", (slug,))
            if not cur.fetchone():
                return slug
            suffix = f"-{i}"
            slug = base_slug[: 180 - len(suffix)] + suffix
            i += 1


def normalize_text(s: str) -> str:
    return re.sub(r"\s+", " ", (s or "")).strip()


def clean_html_fragment(html: str) -> str:
    soup = BeautifulSoup(html or "", "html.parser")
    for tag in soup(["script", "style", "noscript", "iframe", "form", "svg"]):
        tag.decompose()
    return str(soup)


def html_to_markdownish(html: str) -> str:
    soup = BeautifulSoup(html or "", "html.parser")
    lines = []
    for el in soup.find_all(["h1", "h2", "h3", "p", "li"]):
        text = normalize_text(el.get_text(" "))
        if not text:
            continue
        if el.name == "h1":
            lines.append(f"# {text}")
        elif el.name == "h2":
            lines.append(f"## {text}")
        elif el.name == "h3":
            lines.append(f"### {text}")
        elif el.name == "li":
            lines.append(f"- {text}")
        else:
            lines.append(text)
    return "\n\n".join(lines)


def fetch_url(url: str):
    response = requests.get(url, headers=get_random_headers(), timeout=20)
    response.raise_for_status()
    response.encoding = response.apparent_encoding or response.encoding
    return response.text


def extract_article_links(source: dict, html: str) -> list:
    if source.get("direct_article"):
        return [source["url"]]

    soup = BeautifulSoup(html, "html.parser")
    base = source["url"]
    include_keywords = source.get("include_keywords") or ["blog", "article", "post", "news"]
    links = []
    seen = set()
    for a in soup.find_all("a", href=True):
        href = urljoin(base, a["href"])
        parsed = urlparse(href)
        if not parsed.scheme.startswith("http"):
            continue
        if not any(k.lower() in href.lower() for k in include_keywords):
            continue
        # 去掉锚点，减少重复
        href = href.split("#", 1)[0]
        if href in seen or href.rstrip("/") == base.rstrip("/"):
            continue
        seen.add(href)
        links.append(href)
    return links


def fetch_article_content(url: str):
    try:
        html = fetch_url(url)
        soup = BeautifulSoup(html, "html.parser")

        title = ""
        for selector in ["h1", "article h1", "title"]:
            tag = soup.select_one(selector)
            if tag:
                title = normalize_text(tag.get_text(" "))
                title = re.sub(r"\s+[-|].*$", "", title).strip()
                if 3 <= len(title) <= 120:
                    break

        content = ""
        for selector in ["article", "main article", ".post-content", ".article-content", ".content", ".entry-content", "main"]:
            element = soup.select_one(selector)
            if element:
                content = clean_html_fragment(str(element))
                if len(BeautifulSoup(content, "html.parser").get_text(strip=True)) > 300:
                    break

        if not content:
            paragraphs = [p for p in soup.find_all("p") if len(normalize_text(p.get_text())) > 40]
            content = "\n".join(str(p) for p in paragraphs)

        if not title or not content:
            return None, None, None

        cover = get_cover_image(content, url)
        return title, content, cover
    except Exception as e:
        print(f"❌ 抓取文章内容失败 {url}: {e}")
        return None, None, None


def get_cover_image(content: str, base_url: str = ""):
    soup = BeautifulSoup(content or "", "html.parser")
    for img in soup.find_all("img"):
        src = img.get("src") or img.get("data-src")
        if not src:
            continue
        src = urljoin(base_url, src)
        if src.startswith("http://") or src.startswith("https://"):
            return src
    return None


def ensure_admin_user(conn) -> int:
    with conn.cursor() as cur:
        cur.execute("SELECT id FROM users WHERE role='admin' ORDER BY id LIMIT 1")
        row = cur.fetchone()
        if row:
            return row["id"]
        raise RuntimeError("数据库中没有管理员用户，请先启动后端完成初始化")


def create_category(conn, name: str) -> int:
    slug = slugify(name, "category")
    now = datetime.now()
    with conn.cursor() as cur:
        cur.execute("SELECT id FROM categories WHERE slug=%s OR name=%s LIMIT 1", (slug, name))
        row = cur.fetchone()
        if row:
            return row["id"]
        cur.execute(
            """
            INSERT INTO categories (name, slug, description, sort_order, meta, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            """,
            (name, slug, f"{name} 相关文章", 0, None, now, now),
        )
        return cur.lastrowid


def create_tags(conn, tags: list) -> list:
    ids = []
    now = datetime.now()
    with conn.cursor() as cur:
        for tag_name in tags:
            slug = slugify(tag_name, "tag")[:50]
            cur.execute("SELECT id FROM tags WHERE slug=%s OR name=%s LIMIT 1", (slug, tag_name))
            row = cur.fetchone()
            if row:
                ids.append(row["id"])
                continue
            cur.execute(
                "INSERT INTO tags (name, slug, meta, created_at) VALUES (%s, %s, %s, %s)",
                (tag_name, slug, None, now),
            )
            ids.append(cur.lastrowid)
    return ids


def create_post(conn, title: str, html_content: str, category_id: int, tag_ids: list, source_url: str, user_id: int, status="published"):
    title = normalize_text(title)[:255]
    now = datetime.now()
    with conn.cursor() as cur:
        cur.execute("SELECT id FROM posts WHERE title=%s LIMIT 1", (title,))
        if cur.fetchone():
            print(f"⚠️  文章已存在，跳过：{title}")
            return None

        text = BeautifulSoup(html_content, "html.parser").get_text(" ")
        excerpt = normalize_text(text)[:300]
        markdown = html_to_markdownish(html_content) or excerpt
        base_slug = slugify(title, "post")
        slug = unique_slug(conn, "posts", base_slug)
        meta = json.dumps({"source_url": source_url, "cover_image": get_cover_image(html_content, source_url)}, ensure_ascii=False)

        cur.execute(
            """
            INSERT INTO posts
              (user_id, category_id, title, slug, excerpt, content_markdown, content_html,
               status, type, comment_status, view_count, meta, published_at, created_at, updated_at)
            VALUES
              (%s, %s, %s, %s, %s, %s, %s,
               %s, 'post', 'open', 0, %s, %s, %s, %s)
            """,
            (user_id, category_id, title, slug, excerpt, markdown, html_content, status, meta, now, now, now),
        )
        post_id = cur.lastrowid

        for tag_id in tag_ids:
            cur.execute(
                "INSERT IGNORE INTO post_tag_relation (post_id, tag_id, created_at) VALUES (%s, %s, %s)",
                (post_id, tag_id, now),
            )

        print(f"✅ 创建文章：{title}")
        return post_id


def run(limit_per_source: int, only_category: str = ""):
    conn = get_conn()
    created = 0
    skipped = 0
    try:
        user_id = ensure_admin_user(conn)
        for source in ARTICLE_SOURCES:
            if only_category and source["category"] != only_category:
                continue
            print(f"\n📚 来源：{source['name']} ({source['url']})")
            try:
                source_html = fetch_url(source["url"])
                links = extract_article_links(source, source_html)[:limit_per_source]
                print(f"📖 待抓取 {len(links)} 篇")
                for link in links:
                    print(f"🔍 {link}")
                    title, content, cover = fetch_article_content(link)
                    if not title or not content:
                        print("⚠️  内容不足，跳过")
                        skipped += 1
                        continue
                    category_id = create_category(conn, source["category"])
                    tag_ids = create_tags(conn, source["tags"])
                    post_id = create_post(conn, title, content, category_id, tag_ids, link, user_id)
                    if post_id:
                        created += 1
                    else:
                        skipped += 1
                    conn.commit()
                    delay(1)
            except Exception as e:
                conn.rollback()
                print(f"❌ 来源处理失败：{source['name']}: {e}")
            delay(1)
        print(f"\n🎉 完成：新增 {created} 篇，跳过 {skipped} 篇")
    finally:
        conn.close()


def main():
    parser = argparse.ArgumentParser(description="抓取文章并保存到 MariaDB")
    parser.add_argument("--limit", type=int, default=3, help="每个来源最多抓取几篇，默认 3")
    parser.add_argument("--category", default="", help="只抓取指定分类，例如：生活")
    args = parser.parse_args()
    run(max(args.limit, 1), args.category.strip())


if __name__ == "__main__":
    main()
