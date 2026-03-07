# 阶段一+：exe可访问 Site 与 Admin（Windows，无 Docker）

目标：构建通过、自动化测试通过；打包出的 `bbs-go.exe` 启动后可访问：

- Site：`http://127.0.0.1:8082/`
- Admin：`http://127.0.0.1:8082/admin/`

## 前置说明

Go 服务启动后会按“磁盘目录”提供静态资源（不是 embed 内嵌）。当构建产物存在时会优先使用：

- Admin：`./admin/dist/index.html`
- Site：`./site/.output/public/index.html`

## 1. Site 静态化产物

在仓库根目录执行：

```powershell
pnpm -C site install
pnpm -C site run test:stage2
pnpm -C site run build
pnpm -C site run generate
```

预期产物目录：

- `site/.output/public/`

## 2. Admin 构建产物

在仓库根目录执行：

```powershell
pnpm -C admin install
pnpm -C admin run build
```

预期产物目录：

- `admin/dist/`

## 3. 打包 exe 并验证访问

在仓库根目录执行：

```powershell
go test ./...
go build -o .\bbs-go.exe .
.\bbs-go.exe
```

另开一个 PowerShell 窗口执行可用性检查（返回 200 即可）：

```powershell
(Invoke-WebRequest -UseBasicParsing http://127.0.0.1:8082/).StatusCode
(Invoke-WebRequest -UseBasicParsing http://127.0.0.1:8082/admin/).StatusCode
```
