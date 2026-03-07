# UI/UX 问题跟踪（未解决）

本文件用于记录在 UI/UX 讨论中已提出、但尚未解决的问题与决策点。

---

## A. 已复现问题（需要修复）

| ID | 场景 | 问题描述 | 复现步骤（最短） | 预期 | 当前表现 | 相关代码线索 | 状态 |
|----|------|----------|------------------|------|----------|--------------|------|
| UI-001 | 战斗/暂停 | “退出战斗”无真正退出 | 战斗中 → 暂停 → 点“退出战斗” | 返回主菜单/演示态，结束当前战斗 | 观感上等同 R+Z 重置开局 | [ui.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/ui.ts)（btn-pause-exit）、[backend.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/backend/backend.ts)（returnToMenuPressed） | 待定位 |
| UI-002 | 主菜单 | “多人对战”子菜单不明显 | 主菜单观察“多人对战”展开态 | 二级入口视觉层级清晰 | 二级按钮与一级几乎一致 | [ui.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/ui.ts)（menu-multi）、[styles.css](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/styles.css)（menu 样式） | 待设计 |
| UI-003 | 关于/LAN | 旧版出招表/帮助面板残留 | 打开“关于”或“局域网联机” | 不应出现旧 HUD/帮助残影 | 面板底部还能看到旧帮助内容 | [ui.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/ui.ts)（menu-help/pause-help/instruction-card） | 待定位 |
| UI-004 | 多人子菜单 | “本地双人”点击无反应 | 主菜单 → 多人对战 → 本地双人 | 进入本地双人对局 | 视觉上无反应 | [ui.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/ui.ts)（btn-mode-dual 与 canSwitchMode gating） | 待定位 |
| UI-005 | 局域网联机 | 端口手动设置入口缺失 | 打开“局域网联机” | 可手动指定端口（高级项） | 只能随机端口（port=0） | [ui.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/ui.ts)（startQuickHost）、[tauriLanUi.ts](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src/net/tauriLanUi.ts)、[lib.rs](file:///d:/pama1234/pfp/p-2026-01/duel/duel-app/src-tauri/src/lib.rs) | 待修复 |

---

## B. 体验/架构问题（需要讨论后定方案）

| ID | 主题 | 问题/决策点 | 目标 | 备选方向（简述） | 状态 |
|----|------|-------------|------|------------------|------|
| UX-001 | 单人入口 | “单人游戏”点击后是否直接开局，还是先展示出招表/准备页 | 降摩擦与保上手的平衡 | 直接开局 / 开局前一屏 / 首次强制帮助 | 待讨论 |
| UX-002 | 菜单层级 | 一级/二级菜单在视觉与交互上如何区分 | 让用户一眼理解层级 | 二级缩进/弱化/侧滑面板/切页 | 待讨论 |
| UX-003 | LAN大厅 | LAN 页面整体交互与生命周期/事件体系混乱 | 可维护、状态单一来源 | 场景化 UIPhase + 单向数据流 | 待讨论 |
| UX-004 | 退出语义 | “退出战斗”与“重置开局（R+Z）”语义如何区分 | 用户心理模型一致 | Exit→回主菜单；Reset→本模式重开 | 待讨论 |

