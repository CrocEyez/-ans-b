# Ans-b 校园智能问答系统 — PPT 大纲

---

## 1. 项目背景（1页）

- 校园场景痛点：新生入学、日常校园生活中信息获取渠道分散，微信群/公众号/官网查找效率低
- 现有方案不足：传统 FAQ 系统只能关键词匹配，无法理解语义；人工客服响应慢
- 项目目标：构建基于 RAG（检索增强生成）架构的校园智能问答助手，实现知识录入→向量检索→AI 增强回答的完整闭环
- 项目名称：Ans-b，定位为面向学生的校园生活智能问答平台

---

## 2. 成员分工（1页）

| 成员 | 负责模块 | 具体工作 |
|------|----------|----------|
| 蒋睿智 | 智能问答 | 向量检索（pgvector）、Embedding 服务部署、AI 回答生成（OpenAI 兼容 API）、相似度阈值策略 |
| 石家林 | 用户管理 | 学生注册登录、JWT + Redis Session 鉴权、管理员系统、用户列表 API、Console 管理端 |
| 聂梓原 | 知识投稿管理 | 知识录入与向量化、学生投稿提交、管理员审核流程（通过/驳回→入库）、投稿状态管理 |

---

## 3. 相关技术栈（1页）

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| 后端框架 | Go + Gin | 高性能 HTTP 服务，RESTful API |
| 数据库 | PostgreSQL + pgvector | 业务数据存储 + 向量检索 |
| 缓存/会话 | Redis | JWT Session 管理，支持主动失效 |
| Embedding | FastAPI + bge-large-zh-v1.5 | 中文文本向量化，本地部署 |
| AI 大模型 | OpenAI-compatible API | DeepSeek / Kimi，Chat Completions |
| Web 管理端 | Vue 3 + Vite + TDesign | SPA，侧边栏布局 |
| 桌面客户端 | Wails v2 + Vue 3 | Go + Webview，跨平台 |
| 容器化 | Docker Compose | 一键部署全部服务 |
| 鉴权 | JWT + bcrypt + Redis | 无状态令牌 + 服务端会话 |

---

## 4. 开发过程（1页）

```
v0.1 (5/24)  项目初始化、数据库部署、向量问答闭环
v0.2 (5/30)  鉴权体系（JWT+Redis）、Embedding Docker 化
v0.3 (6/02)  Wails 桌面客户端、学生投稿、审核骨架
v0.4 (6/14)  管理端审核完善、审核通过自动入库
v0.5 (7/04)  Docker 一键部署、Console 侧边栏、用户管理、热点问题
```

- 开发周期：2026年5月15日 → 7月4日（约7周）
- 代码规模：33 个核心文件，1724+ 行新增代码
- 协作方式：Git 分支开发 + PR 合并

---

## 5. 需求分析（1-2页）

### 功能性需求
- 学生注册/登录
- 自然语言提问，AI 生成回答
- 学生投稿校园知识
- 管理员审核投稿（通过/驳回）
- 管理员直接录入知识
- 知识库浏览（分页）
- 用户管理
- 热点问题统计
- 相似度不足时不强行回答

### 非功能性需求
- 响应时间 < 3s
- 支持 100+ 并发
- Docker 一键部署
- 跨平台桌面客户端

---

## 6. 系统设计（2-3页）

### 6.1 总体架构图

```
┌──────────────┐   ┌──────────────┐
│ Web Console  │   │ Wails Client │
│  (Vue 3)     │   │  (Vue 3+Go)  │
└──────┬───────┘   └──────┬───────┘
       │ HTTP/JSON         │ HTTP/JSON
       └────────┬──────────┘
                ▼
       ┌────────────────┐
       │  Go Server     │
       │  (Gin + pgx)   │
       └───┬──────┬─────┘
           │      │
     ┌─────▼──┐ ┌─▼──────────┐
     │PostgreSQL│ │ Embedding  │
     │+pgvector │ │ (FastAPI)  │
     └──────────┘ └─────┬──────┘
                        │
                 ┌──────▼──────┐
                 │  OpenAI API │
                 │ (DeepSeek)  │
                 └─────────────┘
```

### 6.2 数据库设计
- `users` — 学生用户
- `admin_users` — 管理员
- `knowledge_items` — 知识条目（含 access_count 热度统计）
- `knowledge_chunks` — 向量分块（1024 维）
- `user_submissions` — 学生投稿（pending/approved/rejected）

### 6.3 RAG 问答流程
```
用户提问 → 问题向量化 → pgvector TopN 检索 → 相似度判断
  ├─ 命中（≥阈值） → 候选知识 + Prompt → AI 生成回答
  └─ 未命中 → 返回候选列表 + 提示无法确认
```

### 6.4 审核流程
```
学生投稿 → pending → 管理员审核
  ├─ 通过 → 向量化 → 写入 knowledge_items + chunks
  └─ 驳回 → 记录原因 → 状态变更
```

---

## 7. 系统实现（2-3页）

### 7.1 智能问答模块（蒋睿智）
- pgvector 向量检索：1024 维 bge-large-zh-v1.5 embedding，余弦相似度排序
- 相似度阈值 0.40，低于阈值不强行回答
- 访问计数：每次命中 `access_count++`，支撑热点问题排行
- AI 增强：拼接候选知识 + 用户问题 → OpenAI Chat Completions → Markdown 解析
- 关键技术：`<=>` 余弦距离算子 + `ORDER BY` + `LIMIT`

### 7.2 用户管理模块（石家林）
- 注册：bcrypt 密码哈希（cost=12）→ PostgreSQL `users` 表
- 登录：JWT 签发（HS256，24h）+ Redis Session 存储（key：`auth:session:{id}`）
- 管理员：启动时自动初始化 `admin/admin123`，可扩展
- Console 用户列表：分页查询 `SELECT ... LIMIT N OFFSET M`，页码跳转
- 鉴权中间件：Bearer Token → JWT 验证 → Session 查询 → 角色检查

### 7.3 知识投稿管理模块（聂梓原）
- 学生投稿：`CreateSubmission` → `user_submissions` 表（status=pending）
- 管理员审核：列表展示（状态筛选）→ 通过（自动向量化 + 写入 knowledge_items/chunks）→ 驳回（备注原因）
- 知识录入：管理员直接填表 → 调用 Embedding API → 写入知识库
- 知识库浏览：分页表格，支持分类/状态/关键词搜索

---

## 8. 难点及解决方案（1-2页）

| 难点 | 解决方案 |
|------|----------|
| Embedding 模型部署 | Docker 容器化 `bge-large-zh-v1.5`，FastAPI 服务，1024 维向量 |
| 中文语义检索精度 | 选用中文优化模型 BGE，相似度阈值调优（0.40），候选 Top5 交 AI 二次判断 |
| JWT 无法主动失效 | Redis Session 存储，登出时删除 key，服务器可控踢人 |
| Docker 镜像国内拉取失败 | 配置 daemon.json 镜像加速 + `pull_policy: if_not_present` |
| Go 模块代理超时 | Dockerfile 内设置 `GOPROXY=https://goproxy.cn,direct` |
| Vue 控制台鉴权 | Admin QA 独立路由 `/api/v1/admin/qa/ask`，控制台传 Bearer Token |
| 数据库字段缺失 | `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 补字段，`IF NOT EXISTS` 建表语法的坑 |
| Wails 跨平台编译 | MinGW-w64 交叉编译工具链 + `wails build -platform windows/amd64` |

---

## 9. 项目总结与展望（1页）

### 已完成
- ✅ RAG 问答完整闭环（知识录入 → 向量检索 → AI 回答）
- ✅ 学生端：注册/登录/问答/投稿/历史/热点
- ✅ 管理端：审核/录入/用户管理/问答测试
- ✅ Docker 一键部署
- ✅ Windows 客户端编译

### 待改进
- 🔜 文档导入（PDF/Word/Markdown 解析）
- 🔜 IVFFlat 索引优化（数据量达标后创建）
- 🔜 对话历史与上下文记忆
- 🔜 反馈机制（回答质量评分）
- 🔜 移动端适配

---

## 10. 系统演示（现场操作）

演示顺序（按脚本 `docs/demo-script.md`）：

1. **学生端登录**：注册 → 登录 → 进入问答页
2. **AI 问答**：提问「食堂几点关门」→ AI 回答 + 候选结果
3. **学生投稿**：填写知识 → 提交 → Toast 提示 → 查看投稿状态
4. **管理端审核**：登录 admin → 进入审核端 → 通过一条投稿 → 驳回一条
5. **管理端知识录入**：直接录入知识 → 保存 → 知识库浏览验证
6. **管理端问答测试**：验证新录入的知识可检索
7. **用户管理**：浏览用户列表 → 分页跳转
8. **热点问题**：切换客户端 → 查看热度排行

---

> 演示视频脚本详见：`docs/demo-script.md`
