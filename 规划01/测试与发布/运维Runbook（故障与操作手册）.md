# 运维 Runbook（故障与操作手册）

本文档沉淀“可照做”的部署、验收、回滚、排障步骤。它不替代 [`部署与运维方案.md`](部署与运维方案.md) 的架构设计，而是作为上线后的操作手册。

---

## 1. 约定与边界

- 本仓库不提交敏感信息（`.env`、密钥、口令等）。
- `server/web_dist/` 为网站构建产物（通常不入库），但会作为部署输入。
- 字体与分包产物体积较大，不入库；构建阶段会自动拉取并生成（详见 [`规划01/平台与设置/字体与样式规范（MapleMono）.md`](../平台与设置/字体与样式规范（MapleMono）.md)）。

---

## 2. 快速部署（最小闭环）

目标：让服务“可访问、可验收、可回滚”。

### 2.1 前置条件

- 服务器：已安装 Docker 与 Docker Compose
- 代码侧：能在本地构建出网站产物 `server/web_dist/`
- 配置侧：目标机存在 `.env` 与 `data/` 目录（或按首次部署流程创建）

### 2.2 本地构建网站产物

原则：部署时不要依赖远端在线构建（便于复现与回滚），将构建产物随发布一并归档。

- 生成字体（若缺失会自动下载/解压/生成）与构建网站：
  - `pnpm -C duel-app run build:website`

产物检查：
- `server/web_dist/index.html` 存在
- `server/web_dist/about.html`、`privacy.html` 存在（用于固定验收 URL）

### 2.3 目标机启动/更新服务

仓库使用 `docker-compose.yml` 编排（见仓库根目录文件）。

- 更新输入：
  - `docker-compose.yml`
  - `server/go/`（Go 服务二进制与静态文件服务逻辑）
  - `server/web_dist/`（网站产物）
- 启动/更新：
  - `docker compose up -d --build`

注意：
- Compose 同时存在 “镜像内 COPY web_dist” 与 “宿主机挂载 web_dist” 两种路径。当前编排会通过 volume 将 `./server/web_dist` 以只读方式挂载进容器（更新网站时替换 web_dist 即可生效）。

### 2.4 远程无交互推送（可选）

当你希望“本地构建 → 打包 → 推送 → 远端重启 → 探针验收”全程无交互时，按以下文档执行：

- [`远程无交互部署与调试流程.md`](远程无交互部署与调试流程.md)

---

## 3. 线上验收（Checklist）

### 3.1 基础可用性

- `GET /api/v1/health`：返回 200
- `GET /`：返回 200，且页面引用 `/assets/*.js`（而不是源码路径）

### 3.2 固定页面存在性（避免 404）

- `GET /privacy`：返回 200 且有正文内容

### 3.3 星历静态资源（只验收“可访问性”）

- `GET /ephem-cache/manifest.json`：返回 200（建议为 `no-cache`）
- `GET /ephem-cache/bodies/*.eph.gz`：返回 200（建议为 `immutable`）

---

## 4. 回滚策略（可执行）

目标：出现严重故障时能在最短路径恢复服务。

### 4.1 回滚输入

建议每次发布都保留一份“部署包”（或可复现输入）：
- `docker-compose.yml`
- `server/web_dist/`（网站产物）
- 以及必要的服务端二进制/镜像版本信息

### 4.2 回滚动作

常用两种回滚方式：
- 回滚网站：恢复上一版 `server/web_dist/`（与 compose 启动无关，属于静态资源回滚）
- 回滚服务：恢复上一版镜像或代码输入后 `docker compose up -d --build`

回滚验收：
- 重跑本 Runbook 的“线上验收（Checklist）”

---

## 5. 常见故障排查（从外到内）

### 5.1 页面 404/内容为空

现象：
- `/privacy` 404 或跳转异常

检查顺序：
- 确认 `server/web_dist/privacy.html` 是否存在（构建产物缺失）
- 确认 Go 静态路由是否包含 `/privacy -> /privacy.html` 映射（路由缺失）
- 确认容器内 `WEBDIR` 指向的目录确实挂载了 `web_dist`（挂载点错/未更新）

### 5.2 首页显示异常但 API 正常

现象：
- `/api/v1/health` 正常，但 `/` 白屏/样式错乱

检查顺序：
- 确认 `server/web_dist/assets/` 下存在对应的 js/css（构建未产出或产物未同步）
- 检查响应头缓存策略：HTML 是否为 `no-cache`，避免旧 HTML 指向不存在的 hash 资源

### 5.3 静态资源缓存导致“更新不生效”

现象：
- 部署后用户端仍看到旧页面或旧数据

处理建议：
- HTML 使用 `no-cache`，hash 资源使用 `immutable`（原则：入口不缓存、内容缓存）
- 对“会变但路径固定”的 JSON/manifest 资源同样使用 `no-cache`（避免 manifest 被缓存导致前端取到旧版本）

### 5.4 字体相关问题（构建时报错/字体缺失）

现象：
- 构建阶段字体生成失败或网页字体回退

处理建议：
- 确认本地可用 Docker（Windows 下字体生成默认走 Docker）
- 重新运行 `pnpm -C duel-app run fonts:prepare` 生成分包，再执行 `build:website`
