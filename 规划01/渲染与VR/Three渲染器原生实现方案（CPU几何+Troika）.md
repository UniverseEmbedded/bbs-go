# Three 渲染器原生实现方案（CPU 几何 + Troika）

本文档定义 duel-app 的 Three 渲染后端应如何实现为“真实 three.js 原生渲染”（WebGL），用于替代当前占位符方案（Canvas2D → CanvasTexture → Plane）。目标是：在保持 BackendSnapshot 不变的前提下，达成与 Pixi/p5 肉眼一致的视觉，并具备可控的性能上限（稳定少量 draw calls、低 GC）。

---

## 0. 范围与非目标

### 0.1 范围

- duel-app 前端渲染后端 `renderer=three` 的实现方式
- 2D 风格画面在 three.js 中的表达（正交相机 + 2D 坐标系）
- “粗线/描边/圆弧/圆环/粒子”在 WebGL 环境下的做法（CPU 生成几何）
- HUD 文本使用 Troika Text（与字体规范一致）

### 0.2 非目标

- 不改 BackendSnapshot 结构、不引入渲染态写回核心状态
- 不在本方案中设计三维特效、后处理管线、VR 视角
- 不把 Pixi/p5 的绘制 API 再抽象一层成为通用“绘图指令 IR”

---

## 1. 设计目标

- 视觉一致：与 Pixi/p5 目标“肉眼无差别”（线宽、alpha、旋转中心、圆弧进度、震动）
- 性能可控：避免“每帧整屏纹理上传”；避免每帧 new 大量对象；稳定少量 draw calls
- 结构可扩展：后续能自然扩展到 VR/WebXR（相机、层级、文本不阻塞）
- 确定性不受影响：渲染只消费快照；粒子仍按快照/seed 重建策略执行

---

## 2. 坐标系与相机（关键决策）

### 2.1 统一坐标系

- 世界坐标直接复用 `INTERNAL_CANVAS_SIDE_LENGTH` 的内部坐标（例如 0..SIDE）
- Snapshot 侧的 Y 轴保持“向下为正”（与 p5/Pixi 一致）

### 2.2 相机配置

使用正交相机把内部坐标直接映射到屏幕：

```ts
const SIDE = INTERNAL_CANVAS_SIDE_LENGTH;
const camera = new THREE.OrthographicCamera(0, SIDE, SIDE, 0, 0.1, 100);
camera.position.z = 10;
```

说明：

- 三维世界的“数学坐标”仍是 Y 向上；为保持渲染 API 与 snapshot 一致，使用一个根节点做屏幕坐标变换：
  - `screenRoot.scale.y = -1`
  - `screenRoot.position.y = SIDE`
- 之后所有“游戏画面实体”（背景线/玩家/箭/粒子/环等）都直接使用 snapshot 的 `x/y/rotation` 驱动，无需再做 `y = SIDE - y` 的逐点转换。

### 2.3 缩放与像素比

- CSS 尺寸仍由外层 `canvasSideLength` 控制
- renderer 仅负责 `setSize(canvasSideLength, canvasSideLength)` 与 `setPixelRatio(dprClamp)`
- 线宽使用“世界单位线宽”并叠加 `lineWeightScale`，随屏幕缩放自然变化

---

## 3. 总体结构（批处理优先）

Three 渲染器由固定数量的“批（Batch）”组成，每个批尽量对应 1 个 Mesh/Material，且生命周期为“初始化一次、每帧只更新 buffer 内容”。

建议批次：

- Batch A：背景线（20 条）— 动态 `BufferGeometry`（粗线 quad）
- Batch B：玩家填充方块（2 个）— `InstancedMesh`（unit quad）
- Batch C：玩家描边（2 个）— CPU 粗线 quad 或 shader 边框（任选其一，优先 CPU 粗线以对齐 p5/Pixi）
- Batch D：箭（多条）— 粗线 quad（箭杆/羽毛）+ 小多边形（箭头）
- Batch E：粒子（多种类型）— 优先分两类：
  - E1：线类/环类粒子：粗线 quad / 分段环
  - E2：点/方块类粒子：`InstancedMesh`（unit quad）
- Batch F：HUD 圆弧（倒计时环等）— 分段弧线（粗线 quad）

约束：

- 禁止每个粒子/箭/线段生成单独 Mesh
- 禁止每帧 `scene.clear()` 重建图元树

---

## 4. 粗线渲染（CPU 生成几何）

### 4.1 线段到 quad

对任意线段 `(x1,y1)-(x2,y2)` 和厚度 `t`，在 CPU 侧生成一个 quad（2 个三角形）：

- 计算方向 `d = normalize(p2 - p1)`
- 计算法线 `n = ( -d.y, d.x )`
- 顶点：
  - `p1 + n * (t/2)`, `p1 - n * (t/2)`, `p2 - n * (t/2)`, `p2 + n * (t/2)`
- index：`0,1,2, 0,2,3`

### 4.2 线帽与 join

本项目当前需求以“独立线段”为主（背景线、箭杆、光束），优先实现：

- 线帽：butt（与 Pixi `LINE_CAP.BUTT` 对齐）
- join：miter 不做通用 join（避免复杂度），多段折线按“段段独立”绘制

若未来出现必须连续 join 的路径，再升级为“polyline join”算法或改用 shader 扩展线。

### 4.3 Buffer 组织

- positions：`Float32Array`，按顶点顺序写入 `x,y,z`
- colors：`Float32Array` 或 `Uint8Array(normalized)`，按顶点写入 `r,g,b,a`
- index：`Uint16Array/Uint32Array`，一次性初始化（按最大段数固定）

更新策略：

- 预分配最大段数 `MAX_SEGMENTS`
- 每帧把实际段数记为 `segmentCount`，设置 `geometry.setDrawRange(0, segmentCount * 6)`
- 仅 `positions/colors` 标记 `needsUpdate = true`

---

## 5. 圆弧与圆环（分段折线 + 粗线）

### 5.1 采样策略

圆弧从 `start` 到 `end`，按固定段数 `N` 采样为折线：

- N 默认 32（可按屏幕尺寸/线宽动态调节）
- 每段作为独立线段输入“粗线 batch”

### 5.2 进度弧

倒计时环（进度 0..1）使用：

- 采样整圈得到 `N` 段
- 仅提交前 `floor(N * progress)` 段（或提交全量段但 drawRange 控制）

---

## 6. 玩家与箭（几何选择）

### 6.1 玩家

- 填充方块：`InstancedMesh(PlaneGeometry(1,1))`，实例 `position/rotation/scale` 与颜色
- 描边：优先用 4 条粗线段（四边）复刻 p5/Pixi 线宽与边角风格

### 6.2 箭

建议拆为：

- 箭杆/羽毛：线段（粗线）
- 箭头：小多边形（2 个三角形即可），按箭类型确定尺寸

---

## 7. 粒子（按类型分组，避免高分支 shader）

粒子数量与类型波动大，优先按“几何表达方式”分组：

- 线/环：进入粗线 batch（环用分段折线）
- 点/方块：进入 `InstancedMesh`（unit quad，alpha/颜色/旋转/缩放由实例数据驱动）

约束：

- 不为每粒子创建 Mesh
- 不在 render() 中创建临时数组（用预分配 typed array 写入）

---

## 8. Troika Text（本版必须落地）

### 8.1 覆盖范围

Three 渲染器内的文字（不再依赖 Canvas2D）至少包括：

- “Go”
- 倒计时数字
- 结果文本（胜负）
- “Press R to reset.”
- Stats（FPS/TPS）若继续由渲染器输出

### 8.2 字体规范与资源来源

- 字体族名与风格必须与 [字体与样式规范（MapleMono）.md](file:///d:/pama1234/pfp/p-2026-01/duel/规划01/平台与设置/字体与样式规范（MapleMono）.md) 一致
- Troika 需要可解析的字体文件（`.ttf/.otf/.woff`，不支持 `.woff2`）
- 字体资源路径在 Web 与 Tauri 下都必须可用；禁止默认依赖公网 CDN

建议策略（实现层面固定为“本地生成 + 忽略入库”，详见字体规范的“字体更新流程”）：

1) `pnpm -C duel-app run fonts:prepare` 会下载/解压字体源码，并生成 website 用的 woff2 分包文件
2) 同时把 Regular/Bold TTF 复制到 `duel-app/public/fonts/maple-mono/` 供 Troika 运行时读取
3) TTF/zip 均不入库（gitignore），开发/CI 需在运行 three 渲染器或跑相关构建前执行 `fonts:prepare`

### 8.3 Troika 的更新与渲染策略

- Text 实例在初始化时创建并加入场景，render() 只更新：
  - `text`, `position`, `fontSize`, `color`, `anchorX/anchorY`, `visible`, `opacity`
- 每帧更新后调用 `text.sync()`（或采用“脏标记”策略：仅 text/size/font 改动时 sync）

### 8.4 文本对齐与坐标翻转（实现细节）

Troika Text 的对齐要与 p5 的 `textAlign(CENTER,CENTER)`/Pixi 的 anchor 行为对齐：

- 使用关键字锚点：`anchorX="center"`, `anchorY="middle"`
- 由于 `screenRoot` 做了 `scaleY=-1`，文本需要再“翻回去”，否则会上下颠倒：
  - 对每个 Text 使用一个父节点 `node.scale.y = -1`
  - render 时只更新父节点的位置（`node.position.set(x,y,0)`），Text 保持本地原点对齐即可

---

## 9. 生命周期与资源释放

- `resize()`：只做 renderer size/pixelRatio 与必要的 viewport 更新；不 dispose/recreate 大对象
- `dispose()`：释放 geometry/material/renderer，移除 domElement

---

## 10. 验收标准（必须可验证）

- 视觉：与 Pixi 目标肉眼无差别（背景线、线宽、箭、粒子、倒计时环、文本布局）
- 性能：不出现“整屏纹理每帧上传”路径；draw calls 稳定；无明显 GC 抖动
- 稳定性：切换渲染器不泄漏 WebGL 资源（多次切换后无持续内存增长）

---

## 11. Shader 策略与自动化保障（实现细节）

### 11.1 为什么避免自写 ShaderMaterial

three.js 的 `tonemapping/colorspace` shader chunk 会随版本变化；在自写 shader 中手动 `#include <colorspace_*>` 很容易出现重复定义，导致片元 shader 编译失败（典型报错为 `LINEAR_SRGB_TO_LINEAR_DISPLAY_P3` 等常量/函数 redefinition），表现为“Troika 文本能显示但其它几何都不显示”。

### 11.2 当前落地策略：MeshBasicMaterial + onBeforeCompile 注入 alpha

- 图元批处理使用 `MeshBasicMaterial({ vertexColors: true, transparent: true })`，复用 three 内置的 `tonemapping/colorspace` 管线
- 通过 `material.onBeforeCompile` 注入：
  - vertex：新增 `attribute float alpha; varying float vAlpha;`，并在 `#include <color_vertex>` 后写入 `vAlpha = alpha;`
  - fragment：在 `#include <opaque_fragment>` 后写入 `gl_FragColor.a *= vAlpha;`

### 11.3 单元测试保障

新增单测固定验证上述 shader patch 能正确注入，确保未来升级 three 时不会静默失效：

- `duel-app/scripts/unit-tests.ts`：`three meshbasic shader can be patched with per-vertex alpha`
