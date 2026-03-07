# 字体与样式规范（MapleMono）

本规范用于统一网站与游戏（含 HUD/联机面板/画布文本）的字体与基础样式，目标是让“所有文本”使用 MapleMono，并且在 Web 与 Tauri 两种运行环境下保持可靠加载与可控回退。

---

## 硬约束

- 所有文本使用 MapleMono（全量字符），不做子集化
- Web 端允许 CDN 优先并提供自有源回退
- Tauri 端必须本地打包字体资源，不访问 CDN

---

## 资源与文件组织建议

为了避免把“Web 与 Tauri 的字体加载策略”混在一份 CSS 里，建议拆为两份：

- `fonts.web.css`：`local()` → CDN → 自有源（同一路径的备用源）
- `fonts.app.css`：只引用本地相对路径（不出现 `https://`）

上层样式统一通过同一个 `font-family` 名称引用，保证 UI/渲染层一致。

---

## Web 端加载策略（CDN 优先 + 自有源回退）

Web 端可以利用 `@font-face src` 的顺序回退能力，按以下顺序尝试：

1. `local()`（用户机器若安装了就直接用）
2. CDN（命中率高）
3. 自有域名/站点资源（CDN 故障时的兜底）

示例（请将 URL 替换为你实际使用的 CDN 与自有源路径）：

```css
@font-face {
  font-family: "Maple Mono";
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src:
    local("Maple Mono"),
    local("Maple Mono CN"),
    url("https://your-cdn.example.com/fonts/MapleMono.woff2") format("woff2"),
    url("/fonts/MapleMono.woff2") format("woff2");
}
```

---

## Tauri 端加载策略（仅本地打包）

Tauri 端应确保字体文件随应用发行包分发，并使用相对路径引用。例如将字体放入前端静态资源目录后，通过：

```css
@font-face {
  font-family: "Maple Mono";
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url("./fonts/MapleMono.woff2") format("woff2");
}
```

Tauri 端禁止默认访问 CDN，以保证离线可用、启动确定性与隐私边界。

---

## 全站字体栈（含 emoji 回退）

以下为推荐的字体栈结构：

- 正文字体与 UI：MapleMono → 系统 sans 回退 → emoji 回退
- 代码字体：等宽字体优先 → MapleMono → 系统等宽回退 → emoji 回退

示例（CSS 变量）：

```css
--font-family:
  "Maple Mono",
  BlinkMacSystemFont,
  Helvetica,
  "DejaVu Sans",
  Arial,
  sans-serif,
  "Apple Color Emoji",
  "Segoe UI Emoji",
  "Noto Color Emoji",
  emoji;

--font-family-code:
  "Maple Mono",
  ui-monospace,
  SFMono-Regular,
  Menlo,
  Monaco,
  Consolas,
  "Liberation Mono",
  monospace,
  "Apple Color Emoji",
  "Segoe UI Emoji",
  "Noto Color Emoji",
  emoji;
```

---

## 渲染层对齐要求

- DOM/HUD：默认继承 `--font-family`
- p5 文本：必须显式指定与 CSS 相同的字体族名，避免不同平台默认字体造成 `textWidth` 与排版差异
- Pixi 文本：必须显式指定 `fontFamily: "Maple Mono"`，并统一字体权重/大小策略
- three.js/Troika Text：字体资源选择与加载必须以本规范为准，避免在渲染迁移计划中另起一套字体方案

---

## 字体更新流程（源码不入库）

由于 MapleMono 的 TTF/WOFF2 文件较大，仓库中将字体源文件与生成的 woff2 分包文件统一做 gitignore，不直接提交到仓库。

### 下载字体源码

- 字体来源：MapleMonoNormal-NF-CN-unhinted.zip（GitHub Release）
  - https://github.com/subframe7536/maple-font/releases/download/v7.9/MapleMonoNormal-NF-CN-unhinted.zip
- 下载与解压目录：`tools/fonts/MapleMonoNormal-NF-CN-unhinted/`（已在 .gitignore 中忽略）

### 生成网页用分包字体（woff2 + result.css）

在仓库根目录执行：

```bash
pnpm -C duel-app run fonts:prepare
```

说明：
- 该命令会在缺少分包产物时自动下载并解压字体 zip，再用 `cn-font-split` 生成 woff2 分包与 `result.css`。
- Windows 环境下默认使用 Docker 容器执行 `cn-font-split`，以避免本地原生依赖缺失问题。
- 生成目录：`server/web_page/fonts/maple-mono/`（woff2 与 ttf 均被忽略）
- 同时会把 Regular/Bold 的 TTF 复制到 `duel-app/public/fonts/maple-mono/`，用于 three.js/Troika Text 运行时加载（TTF 不入库，已 gitignore）

注意：
- Troika Text 不支持 `.woff2`，因此不能直接复用 `cn-font-split` 生成的分包产物；three.js 渲染器应读取上述前端静态目录中的 `.ttf`。

### 构建网站

```bash
pnpm -C duel-app run build:website
```

该命令会自动先执行 `fonts:prepare`，确保构建时字体分包文件存在。

