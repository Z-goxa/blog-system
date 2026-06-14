<template>
  <div class="about">
    <h2>关于这个博客</h2>

    <p class="about-lead">
      这里是景龙的个人博客。写代码、搭工具、折腾新东西——边学边记，边做边写。
    </p>

    <p>
      这个博客最早源于一个简单的想法：把平时踩过的坑、解决过的问题、研究过的东西系统地沉淀下来。
      很多东西当时觉得记住了，过几个月就模糊了，不如写下来。
      如果同时能帮到其他开发者，那就更好了。
    </p>

    <p>
      内容方面比较杂，技术类的偏多，但也有些生活随笔、读书笔记和工具评测。
      技术向的主要集中在 Go、Web 开发、开源项目、桌面应用和一些系统架构相关的东西。
    </p>

    <h3>关于这个站</h3>

    <p>
      前端用的 Vue 3 + Vite，后端是 Go + GORM + MariaDB。
      整体走轻量路线——没有复杂的微服务、没有云原生编排、不追时髦，够用就行。
      样式主题参考自 xgxa 的布局风格，简洁为主。
    </p>

    <p>
      所有文章支持 Markdown 编辑（goldmark 渲染，GFM 语法），
      支持分类归档和标签索引，访客可以普通搜索，评论需要注册账号。
    </p>

    <h3>联系我</h3>

    <p>
      文章里如果有错误或者有更好的做法，欢迎在评论区指出来，一起讨论。
    </p>

    <p class="about-signature">
      景龙<br>
      2026 年 6 月<br>
      于上海
    </p>
  </div>
</template>

<style scoped>
.about { font-size: 15px; line-height: 1.85; color: #333; max-width: 680px; }
.about h2 { font-size: 26px; font-weight: 700; color: #222; margin: 0 0 22px; }
.about h3 { font-size: 17px; font-weight: 700; color: #222; margin: 2em 0 0.8em; padding-bottom: 6px; border-bottom: 2px solid #f0f0f0; }
.about p { margin: 0.8em 0; }
.about-lead { font-size: 16px; color: #444; }
.about-signature { margin-top: 2.5em; color: #999; font-size: 14px; }
.about code {
  background: #f7f7f7; padding: 2px 5px;
  font-family: 'Menlo', 'Consolas', monospace;
  font-size: 0.9em; color: #c7254e; border: 1px solid #ededed;
}
</style>
