# LAN 发现与端内开服设计

本文档定义 LAN 自动发现（UDP 广播 + mDNS）与端内开服（同进程后台任务）设计，用于桌面与安卓端（Tauri）。纯 Web 客户端不具备开服能力，仅作为“发现/连接/输入/渲染”的前端。

---

## 1. 目标与非目标

### 1.1 目标

- 支持端内开服：在同一 App 进程内启动 WebSocket 服务端（不拉起独立进程）。
- 支持自动发现：以 mDNS 为主路径；UDP 广播作为可选回退/加速；并做去重与过期处理。
- 支持手动兜底：UI 始终保留手填 wsUrl 的连接方式。
- 支持房间识别码：识别码按房间独立（3 位数字），加入房间时校验；错误尝试触发等待窗口以抑制穷举。
- 支持单 Server 多房间：发现阶段只发现 Server 入口；房间列表通过连接后拉取（protocol=2）。

### 1.2 非目标（暂不讨论）

- 远程联机（公网/NAT 穿透）与真正的账号/房间系统。
- 自动重连与断线续局。

---

## 2. 术语

| 名称 | 说明 |
|------|------|
| Host | 端内开服的一方（服务端与客户端同机） |
| Guest | 连接到 Host 的客户端 |
| Discovery | 自动发现机制（UDP + mDNS） |
| HostId | Host 实例唯一标识（随机 UUID） |
| HostToken | 主机管理 token（仅主机可创建/管理房间） |
| RoomCode | 房间识别码（三位数字字符串 "000".."999"） |

---

## 3. 端内开服（Host Task）

### 3.1 宿主能力边界

- 桌面与安卓端：由 Tauri Rust 后端启动 HostTask（后台任务/线程），生命周期由前端 UI 控制。
- 纯 Web：只能连接/发现，不能启动 HostTask。

### 3.2 HostTask 子任务

| 子任务 | 作用 |
|--------|------|
| WsListener | 监听 WebSocket，处理 hello/join/input_state/ping 等消息 |
| TickLoop | 固定步长推进后端并广播 snapshot |
| UdpAnnouncer | 周期性广播 HostAnnounce |
| MdnsAdvertiser | 发布 mDNS 服务（TXT 记录带参数） |
| HostStatus | 对前端提供“当前端口、可分享地址、房间列表摘要、HostToken 是否可用”等状态 |

### 3.3 HostTask 生命周期

- start：创建 HostId，选定 ws 监听端口，启动上述子任务。
- stop：停止广播与 mDNS，关闭 ws listener，断开现有客户端连接，释放资源。
- restart：等价 stop → start（HostId 变化）。

---

## 4. 统一发现模型（HostAnnounce）

自动发现的核心是让客户端获得“可连接目标”的最小信息。UDP 与 mDNS 共享同一信息模型，来源不同但内容等价。

### 4.1 HostAnnounce 数据结构

```ts
type HostAnnounce = {
  magic: "GEOM_DUEL_LAN";
  protocol: 2;
  hostId: string;          // UUID
  serverName: string;      // UI 展示名
  wsPort: number;          // WebSocket 端口
  rooms: true;             // protocol=2 的能力标记
  tsMs: number;            // Host 本地时间戳（仅用于调试/排序）
};
```

### 4.2 客户端发现列表条目（聚合后）

```ts
type DiscoveredHost = {
  hostId: string;
  serverName: string;
  wsUrl: string;                 // 由 ip + wsPort 组合
  protocol: 2;
  rooms: true;
  lastSeenMs: number;
  source: "udp" | "mdns" | "both";
};
```

展示字段最小集合：`serverName`、`wsUrl`。

说明：

- `source` 仅用于诊断，不建议作为面向用户的主信息展示。
- 房间识别码是“按房间”而非“按 Server”，不在发现阶段传播。

---

## 5. UDP 广播发现（可选回退）

### 5.1 端口与地址

- UDP 端口：`32123`（固定）
- 发送目标：对可用网卡发送到该网卡的广播地址；若无法获取广播地址，则退化为 `255.255.255.255:32123`
- 接收：绑定 `0.0.0.0:32123` 监听

### 5.2 广播报文

- 编码：UTF-8 JSON
- 单 datagram 一条 JSON，不做换行拼接

示例：

```json
{
  "magic": "GEOM_DUEL_LAN",
  "protocol": 2,
  "hostId": "bdbf2e33-4dbf-45b4-9f97-1ce60e1b0c2a",
  "serverName": "LAN Host",
  "wsPort": 9001,
  "rooms": true,
  "tsMs": 1730000000000
}
```

### 5.3 频率与过期

- Host 广播频率：每 750ms 发送一次（桌面/安卓同一策略）
- 客户端过期阈值：`lastSeenMs` 超过 3000ms 未刷新则判离线并从列表移除

### 5.4 去噪规则

客户端接收 UDP 报文时：

- `magic` 不匹配：丢弃
- `protocol` 不匹配：丢弃
- `wsPort` 非法：丢弃
- `hostId` 非法 UUID：丢弃

---

## 6. mDNS 发现（主路径，必做）

### 6.1 服务类型与实例命名

- 服务类型：`_geomduel._tcp.local.`
- 实例名：`{serverName} ({hostIdPrefix})`，其中 `hostIdPrefix` 为 hostId 的前 8 位（仅用于区分）
- 端口：`wsPort`

### 6.2 TXT 记录

TXT 记录键值建议（均为字符串）：

| key | value | 说明 |
|-----|-------|------|
| magic | GEOM_DUEL_LAN | 过滤噪声 |
| protocol | 2 | 协议版本 |
| hostId | UUID | 与 UDP 一致 |
| serverName | string | UI 展示名（可与实例名不同） |
| rooms | 1 | 支持房间目录（protocol=2） |

### 6.3 解析规则

客户端浏览 mDNS 服务时：

- 只接受 `magic=GEOM_DUEL_LAN` 且 `protocol=2`
- `hostId` 缺失则丢弃（用于去重与稳定性）
- 以 SRV 记录提供的主机名解析出 IP（IPv4/IPv6 均可），组合 `wsUrl`

---

## 7. 智能回退与去重

### 7.1 并行策略

- mDNS 为主路径并持续运行
- UDP 广播作为可选回退/加速：可并行运行；不可用时不影响主流程
- 任一来源发现到条目即可即时展示（用于更快出现列表）

### 7.2 去重策略

- 主键：`hostId`
- 若 `hostId` 为空（不应发生）：退化为 `wsUrl`
- 若同一 `hostId` 同时来自 UDP 与 mDNS：`source="both"`

### 7.3 回退策略

客户端 UI 提示策略：

- 启动发现后 2000ms 内无任何条目：提示“未发现房间，可尝试手动输入 wsUrl”
- 发现运行持续可用，不自动停止
- 用户手动输入 wsUrl 并连接成功后，发现仍可继续（不冲突）

---

## 8. 房间识别码（RoomCode，按房间独立）

RoomCode 的协议细节见 [联机协议规格v2.md](./联机协议规格v2.md)。本节只定义产品与安全预期：

- RoomCode 为三位数字字符串 `"000"`..`"999"`，按房间独立。
- 房间可选择启用/禁用识别码：
  - 启用时：加入该房间必须输入正确的 RoomCode。
  - 禁用时：加入不需要识别码。

### 8.1 错误与等待

识别码错误应触发等待窗口（如 5 秒）以抑制穷举尝试；等待窗口按连接（WebSocket 会话）计数，连接断开后清空。
- code 错误：服务端拒绝加入，并要求等待 5000ms 才允许再次尝试
- 等待维度：按连接（WebSocket 会话）计数；连接断开后计数清空
- 行为目标：抑制同一连接内的高频穷举，不试图提供强安全性

---

## 9. UI 交互建议（LAN 面板）

UI 细节规范以 [联机UI规格-LAN开服与房间列表.md](./联机UI规格-LAN开服与房间列表.md) 为准。本节仅给出最小交互建议。

### 9.1 面板结构

- 服务器发现列表（Server 入口）
  - 条目：serverName / wsUrl / 新鲜度 /（可选）诊断 source
  - 操作：进入服务器 → 拉取房间列表
- 手动连接
  - wsUrl 输入框（保留）
  - 操作：连接后进入房间列表
  - 说明：识别码按房间独立，不在“连接服务器”阶段输入

### 9.2 Host 面板

- 启动/停止 Host
- 显示 wsUrl（包含本机 IP 与端口）
- 房间管理（仅主机可创建/管理房间）
  - 创建房间：可设置是否启用识别码与 roomCode（或自动生成）
  - 房间列表：展示占槽与识别码状态，支持复制邀请
