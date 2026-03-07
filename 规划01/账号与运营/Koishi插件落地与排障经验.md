# Koishi 插件落地与排障经验（脱敏）

本文整理一次在 Koishi 工作区开发、打包、安装插件，以及 Koishi Desktop 排障过程中沉淀的经验要点。文中所有本机路径、账号名等均用占位符替代。

## 术语与目录约定

- `<DUEL_ROOT>`：本仓库根目录
- `<KOISHI_WORKSPACE>`：本仓库内 Koishi 工作区目录（示例：`<DUEL_ROOT>/koishi`）
- `<PLUGIN_DIR>`：插件目录（示例：`<KOISHI_WORKSPACE>/plugins/duel-gate-invite`）
- `<KOISHI_DESKTOP_INSTANCE_DIR>`：Koishi Desktop 实例目录（示例：`%APPDATA%/Koishi/Desktop/data/instances/default`）

## 核心结论

- `koishi/` 目录是一个 Koishi 机器人工作区（workspace/monorepo），不是“一个插件包”。插件应当作为工作区中的单独 package 存放与构建。
- `koishi/external/` 在很多模板里是“外部代码区”，常见做法是被 `.gitignore` 整体忽略；自研插件不要放这里，优先放 `koishi/plugins/*`。
- 脚手架生成的插件有时会自带 `.git/`（嵌套仓库），会影响主仓库追踪；需要将插件并入主仓库时应删除插件目录内的 `.git/`。
- Koishi Desktop 的“终端”默认工作目录可能位于安装目录（如 Program Files）。该目录通常不可写，不能在这里执行 `npm i` 安装依赖或插件。
- Koishi Desktop 的 “Already Running” 可能是锁文件残留或 PID 复用导致的假阳性；WebUI 访问失败也可能是端口被其他软件（如 WSL/Docker）占用。

## 版本选择建议（带日期）

结论：
- 对“开发/维护插件”的场景，不建议优先选择 Windows 桌面版作为主运行环境；优先考虑基于 npm 依赖安装或 Docker 镜像部署，以减少版本滞后带来的兼容问题。
- 桌面版更适合快速上手、可视化管理与临时排障，但其内置的 Koishi 版本与依赖组合通常有自己的发布节奏，不保证与 npm 最新同步。

记录（截至 2026-02-28，按页面显示的发布日期/更新时间）：
- Koishi Desktop：v1.1.3（May 31 / Jun 1, 2024，受时区显示影响）；该版本内置 Koishi 4.17.7。
- Koishi GitHub Releases：Koishi 4.17.12（Aug 15, 2024）。
- npm：koishi 4.18.11（页面显示 Published 18 hours ago）。
- Docker Hub：koishijs/koishi（页面显示 Updated about 9 hours ago）。

## 常见问题与处理

### 1) 插件目录被 git 忽略

现象：插件生成在 `koishi/external/*`，Git 显示被忽略。

原因：工作区 `.gitignore` 里包含 `external`，导致整个目录树被忽略。

建议：
- 将插件移动到 `koishi/plugins/<plugin>`（推荐）。
- 如果确实要留在 `external`，需要在 `.gitignore` 中为特定目录做反忽略规则（但不推荐作为默认方案）。

### 2) `yarn setup` 报 `EEXIST`（mkdir 已存在）

现象：执行脚手架创建插件时提示 `EEXIST: file already exists, mkdir .../src`。

原因：目标目录已存在（通常是之前创建了一半或手动创建过）。

处理：
- 删除目标目录后重新执行脚手架创建命令；或换一个新目录名。

### 3) 插件目录“自成一个 git 仓库”

现象：插件目录下存在 `.git/`，导致主仓库对该目录表现异常（如被识别为嵌套仓库、无法正常加入主仓库变更）。

处理：
- 若插件应纳入主仓库版本控制：删除 `<PLUGIN_DIR>/.git/`。
- 若插件要作为子模块/独立仓库管理：应改用 git submodule/子仓库规范流程（另行规划）。

### 4) Koishi Desktop 提示 “Already Running” 但找不到托盘 / WebUI 不通

可能原因 A：真的还在后台运行
- 处理：检查系统托盘（含隐藏图标 `^` 展开），或在任务管理器查找 Koishi/`koi.exe` 进程并结束。

可能原因 B：锁文件残留 + PID 复用（常见）
- 处理：确保 Koishi Desktop 已退出后，删除锁文件目录中的 `daemon.lock` / `tray.lock`，再重新启动。
- 锁文件通常位于：`%APPDATA%/Koishi/Desktop/data/lock/`

可能原因 C：端口占用导致 WebUI 不可用
- 处理：确认实际监听端口；如果默认端口（如 5140）被 WSL/Docker 等占用，需改端口或使用端口范围内的其他端口。

### 5) 在 Koishi Desktop 安装目录里 `npm i` 报 `EPERM`

现象：在类似 `C:/Program Files/Koishi/Desktop` 目录执行 `npm i <plugin.tgz>` 报 `EPERM mkdir .../node_modules`。

原因：Program Files 目录默认无写权限，npm 无法创建 `node_modules`。

正确做法：
- 切换到实例目录再安装：`cd /d "<KOISHI_DESKTOP_INSTANCE_DIR>"`，然后执行 `npm i <plugin.tgz>` 或 `yarn add <plugin.tgz>`。

补充建议（避免“装完起不来”）：
- Koishi Desktop 实例通常使用 Yarn 启动（取决于实例的 `packageManager` 与 `.yarn/`），因此不要在实例目录混用 `npm i` 与 `yarn`。
- 若确实用了 `npm i` 修改依赖，需再执行 `yarn install` 更新 lockfile，否则实例可能报错并直接退出。

### 6) 依赖已安装，但「插件配置」里搜不到插件

现象：
- 在「依赖管理」页面能看到 `koishi-plugin-xxx` 已安装，但在「插件配置」页面找不到对应条目，也无法搜索到。

原因：
- 「依赖管理」只负责安装依赖到 `node_modules`，不会自动把插件写入 `koishi.yml` 的 `plugins:` 配置树。
- 当插件并未加入 `koishi.yml` 时，Koishi 启动不会加载该插件，因此控制台也不会出现对应的“已添加/可配置”项。
- 若控制台的插件市场/索引不可用（例如出现 Not Found），通过界面选择/检索插件的能力可能进一步受限。

处理（推荐）：
- 直接编辑实例的 `koishi.yml`，将插件加入 `plugins:` 并填写最小配置，然后重启实例。
- Koishi Desktop 实例的配置文件通常位于：`<KOISHI_DESKTOP_INSTANCE_DIR>/koishi.yml`

示例（将插件加到 group:basic）：

```yml
plugins:
  group:basic:
    duel-gate-invite:
      apiBaseUrl: "https://<your-api>"
      botId: "<botId>"
      botSecret: "<botSecret>"
      botPlatform: "onebot"
      botSelfId: "<selfId>"
      group1Id: "<group1>"
      group2Id: "<group2>"
```

验证方法：
- 重启后查看实例日志，确认出现类似 `apply plugin duel-gate-invite:...` 的加载记录。

### 7) OneBot retcode 判读：1400/1200 的常见原因

#### 7.1 `get_group_member_list` retcode 1400 且 args 为空

现象：
- 执行“列举成员/计算差集”相关逻辑时，OneBot 返回 `retcode: 1400`，日志里还能看到请求参数为 `args: {}`。

原因：
- OneBot 侧认为请求参数不合法（典型是缺少 `group_id`），从而直接返回参数错误。
- 在 Koishi 插件场景里，最常见的触发原因是插件配置未填写或填写错误（例如 `group1Id/group2Id` 为空）。

处理：
- 优先检查插件配置的群号是否填写且为正确群号；必要时用引号写成字符串（例如 `group1Id: "123456"`）。
- 确认机器人确实在目标群内，且 OneBot 实现允许拉取群成员列表。

#### 7.2 `send_private_msg` retcode 1200：常见为发送超时/会话不可达

现象：
- 调用 OneBot `send_private_msg` 发送私聊失败，返回 `retcode: 1200`（或其他非 0 失败码），日志表现为发送失败。

原因（经验）：
- 在 NapCat/QQNT 的实现中，常见对应“发送超时”或发送链路未在超时内完成。
- 对未建立会话/临时会话不可达的目标，机器人主动私聊更容易失败；当目标先主动发过消息（建立会话）后，再发送成功率明显上升。

处理建议：
- 先用“自己账号”做一次私聊发送验证，确认链路可用；再对目标测试。
- 对陌生人/未建立会话的目标：先引导对方主动私聊机器人发一句话，再进行后续私聊推送。
- 若问题频繁出现，结合 NapCat 运行日志判断是否为超时/卡顿，并考虑降低并发与限速。

## 推荐工作流（开发 → 打包 → 安装到独立 Koishi）

### 1) 在仓库内开发

- 插件放到：`<KOISHI_WORKSPACE>/plugins/<plugin>`
- 工作区执行依赖安装：`yarn install`
- 开发/调试：启动工作区 Koishi（例如 `yarn dev`），并在 `koishi.yml` 或控制台启用插件。

### 2) 产出可安装包（npm tarball）

目标：得到 `koishi-plugin-xxx-x.y.z.tgz`，可用于离线安装或发布前验证。

- 在插件目录执行：`npm pack`
- `npm pack` 会触发 `prepack`，先构建 `lib/`（JS + d.ts），再打出 tgz。

### 3) 安装到独立 Koishi / Koishi Desktop

- 独立 Koishi 项目：在项目根目录执行 `npm i <path-to-tgz>`，然后在配置中启用插件。
- Koishi Desktop：必须在 `<KOISHI_DESKTOP_INSTANCE_DIR>` 执行安装命令（不要在安装目录里装）。

## 最小命令清单（可复制）

以下命令均为示例，路径请替换占位符。

- 打包 tgz：
  - `cd <PLUGIN_DIR>`
  - `npm pack`
- 在 Koishi Desktop 实例目录安装 tgz：
  - `cd /d "<KOISHI_DESKTOP_INSTANCE_DIR>"`
  - `npm i "<path-to>/koishi-plugin-xxx-x.y.z.tgz"`
