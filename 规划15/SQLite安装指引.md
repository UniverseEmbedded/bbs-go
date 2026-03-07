# SQLite 安装指引（不再出现 /install 引导页）

目标：使用 SQLite 完成后端安装（写入 `bbs-go.yaml` + 创建 `bbs-go.db`），让站点不再被前端强制跳转到 `/install`。

## 你看到 /install 的原因

前端全局中间件会拉取 `/api/config/configs`，当返回的 `installed=false` 时会自动 `navigateTo('/install')`。

只要把后端安装完成（`installed=true`），站点就不会再跳转到 `/install`。

## 步骤（推荐：一键脚本）

1. 在项目根目录启动后端（保持窗口不要关）：

```powershell
.\bbs-go.exe
```

2. 另开一个 PowerShell，在同一目录执行一键安装脚本：

```powershell
.\规划15\一键安装SQLite.ps1
```

3. 安装脚本提示完成后，关闭并重新启动 `bbs-go.exe`。

4. 验收：

- Site：`http://127.0.0.1:8082/` 能正常进入首页，不再跳 `/install`
- Admin：`http://127.0.0.1:8082/admin/`

## 产物说明

安装成功后，项目根目录会出现：

- `bbs-go.yaml`：配置文件（含 `installed: true`、语言、DB 等）
- `bbs-go.db`：SQLite 数据库文件
- `logs/bbs-go.log`：日志文件（若目录存在且开启写入）
