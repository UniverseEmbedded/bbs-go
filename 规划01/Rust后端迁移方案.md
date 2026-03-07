# Rust后端迁移方案

本文档定义将TS后端迁移到Rust的方案，包括duel-core库和duel-server CLI服务端。

---

## 核心决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 数值类型 | f64 | 与TS number对齐，回归更易对齐 |
| 传输协议 | WebSocket | 便于前端接入，开发成本低 |
| 传输格式 | JSON文本帧 | 调试友好，后续可换msgpack |
| 架构 | duel-core + duel-server | 库与二进制分离，可复用 |

---

## 目录结构

```
duel-app/src-tauri/
  Cargo.toml              # workspace root + tauri app package
  src/
    lib.rs
    main.rs
  crates/
    duel-core/
      Cargo.toml
      src/
        lib.rs
        types.rs          # 类型定义
        rng.rs            # 随机数生成器
        backend.rs        # DuelBackend核心
        math.rs           # 数学工具函数
        particles.rs      # 粒子系统
        ai.rs             # AI决策
        tests/
          regression.rs   # 回归测试
    duel-server/
      Cargo.toml
      src/
        main.rs           # CLI入口
        args.rs           # 命令行参数
        protocol.rs       # WS消息定义
        server.rs         # tick loop + 连接管理
        net.rs            # WS工具函数
        tests/
          ws_smoke.rs     # WS冒烟测试
```

---

## duel-core模块设计

### types.rs

对齐TS `backend/types.ts`：

```rust
use serde::{Serialize, Deserialize};

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub enum PlayerId { P1, P2 }

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub enum ControlMode { Single, LocalDual }

#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub enum ArenaLayout { Vertical, Horizontal }

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct PlayerInput {
    pub move_x: f64,
    pub move_y: f64,
    pub aim_x: f64,
    pub shot_pressed: bool,
    pub long_shot_pressed: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct FrontendActions {
    pub reset_pressed: bool,
    pub toggle_help_pressed: bool,
    pub return_to_menu_pressed: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct InputFrame {
    pub tick: u32,
    pub players: [PlayerInput; 2],
    pub actions: FrontendActions,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct RgbColor { pub r: u8, pub g: u8, pub b: u8 }

// PlayerState, ArrowState, GameState, BackendSnapshot 等
// 字段名/含义与TS一致
```

### rng.rs

复刻TS `createSeededRng`：

```rust
pub struct Rng {
    t: u32,
    x: u32,
}

impl Rng {
    pub fn new(seed: u32) -> Self {
        Self { t: seed, x: seed }
    }

    pub fn next_u32(&mut self) -> u32 {
        self.t = self.t.wrapping_add(0x6d2b79f5);
        let mut x = self.x;
        x = x.wrapping_mul(x ^ (x >> 15));
        x ^= x.wrapping_add(x.wrapping_mul(x ^ (x >> 7)) | 61);
        self.x = x ^ (x >> 14);
        self.x
    }

    pub fn random01(&mut self) -> f64 {
        self.next_u32() as f64 / 4294967296.0
    }
}
```

### backend.rs

复刻TS `DuelBackend`：

```rust
pub struct DuelBackend {
    rng: Rng,
    state: BackendState,
}

impl DuelBackend {
    pub fn new(seed: u32) -> Self;
    pub fn new_game(&mut self, demo: bool, instruction: bool);
    pub fn set_control_mode(&mut self, mode: ControlMode);
    pub fn set_arena_layout(&mut self, layout: ArenaLayout);
    pub fn get_snapshot(&self) -> BackendSnapshot;
    pub fn step(&mut self, frame: InputFrame) -> BackendSnapshot;
}
```

状态机：demo → ready → start → play → result

### math.rs

数学工具函数：

- `clamp(value, min, max)`
- `dist(x1, y1, x2, y2)`
- `angle_to(x1, y1, x2, y2)`
- 常量：PI、TWO_PI

### particles.rs

粒子系统：

- Particle结构
- ParticleSet管理
- 生成：ring、爆炸碎片等（type 0/1/2/3）

### ai.rs

AI决策：

- ComputerAIState状态
- aiDecide(player_id)函数

---

## duel-server模块设计

### args.rs

CLI参数（使用clap）：

```rust
pub struct Args {
    #[arg(long, default_value = "0.0.0.0")]
    pub bind: String,

    #[arg(long, default_value = "9001")]
    pub port: u16,

    #[arg(long)]
    pub seed: Option<u32>,

    #[arg(long, default_value = "vertical")]
    pub layout: String,

    #[arg(long, default_value = "info")]
    pub log_level: String,
}
```

### protocol.rs

WS消息定义（与TS对齐）：

```rust
#[derive(Serialize, Deserialize)]
#[serde(tag = "t")]
pub enum Msg {
    Hello { protocol: u32, client: ClientInfo },
    HelloAck { protocol: u32, server: ServerInfo },
    Join { room: String, as: String, name: Option<String> },
    JoinAck { you: String, seed: u32, tick: u32 },
    JoinReject { reason: String },
    InputState { seq: u32, player: String, input: PlayerInput },
    Input { tick: u32, player: String, input: PlayerInput }, // legacy: server treats as InputState
    Snapshot { tick: u32, snapshot: BackendSnapshot },
    Ping { client_time_ms: u64 },
    Pong { client_time_ms: u64, server_time_ms: u64 },
    Kick { reason: String },
    Error { message: String },
}
```

### server.rs

核心结构：

```rust
pub struct RoomState {
    backend: DuelBackend,
    tick: u32,
    last_inputs: [PlayerInput; 2],
    last_input_seq: HashMap<PlayerId, u32>,
    clients: [Option<Client>; 2],
    seed: u32,
}
```

Tick loop：

1. tick += 1
2. 组InputFrame { tick, players: last_inputs, actions: default }
3. snapshot = backend.step(frame)
4. 传输层裁剪：snapshot.particles = []（客户端本地重建粒子）
5. 每 tick 广播 {t: "snapshot", tick, snapshot}

连接处理：

- join时占位p1/p2，满了reject
- 断开时释放slot
- 断开任一玩家可结束当前对局

### net.rs

WS工具函数：

- 接受连接
- 广播消息
- 心跳检测

---

## 回归测试对齐

### 测试策略

Rust core必须与TS core行为一致，使用现有回归基线验证：

1. 使用相同的seed
2. 使用相同的输入序列
3. 生成相同的payload
4. 计算相同的fnv1a32 hash

### regression.rs

```rust
#[test]
fn regression_matches_ts_baseline() {
    let seed = 12345;
    let mut backend = DuelBackend::new(seed);
    backend.new_game(false, false);

    // 与TS runBackendRegression相同的输入序列
    for i in 0..(60 * 12) {
        let frame = InputFrame {
            tick: i as u32,
            players: [
                PlayerInput {
                    move_x: if i % 2 == 0 { 1.0 } else { -1.0 },
                    move_y: 0.0,
                    shot_pressed: i % 15 == 0,
                    long_shot_pressed: i >= 120 && i < 180,
                },
                PlayerInput::default(),
            ],
            actions: FrontendActions::default(),
        };
        backend.step(frame);
    }

    let snapshot = backend.get_snapshot();
    let payload = build_regression_payload(&snapshot);
    let hash = fnv1a32(&payload);

    // 与regression-baseline.json对比
    assert_eq!(hash, EXPECTED_HASH);
}
```

### rng测试

```rust
#[test]
fn rng_matches_ts_values() {
    let mut rng = Rng::new(12345);
    // 与TS打印的预期值对比
    assert!((rng.random01() - 0.12345).abs() < 1e-10);
    // ...
}
```

---

## WS冒烟测试

### ws_smoke.rs

```rust
#[tokio::test]
async fn ws_smoke_test() {
    // 启动server绑定127.0.0.1:0
    let server = start_test_server().await;
    let addr = server.local_addr();

    // 建两个客户端连接
    let mut p1 = connect(addr).await;
    let mut p2 = connect(addr).await;

    // hello (v2 protocol)
    p1.send(Msg::Hello { protocol: 2, client: ClientInfo::default(), host_token: None }).await;
    p2.send(Msg::Hello { protocol: 2, client: ClientInfo::default(), host_token: None }).await;

    // 创建房间并加入（v2 多房间模型）
    p1.send(Msg::CreateRoom { name: "test".into(), code_required: false, room_code: None }).await;
    // 收到 create_room_ack 后获取 roomId
    // p1.send(Msg::JoinRoom { room_id, as: "p1", name: None, room_code: None }).await;
    // p2.send(Msg::JoinRoom { room_id, as: "p2", name: None, room_code: None }).await;

    // 发送输入，确保能收到snapshot
    for tick in 0..120 {
        p2.send(Msg::InputState { room_id: room_id.clone(), seq: tick, input: PlayerInput::default() }).await;
        if let Some(msg) = p2.recv_timeout(Duration::from_millis(100)).await {
            if let Msg::Snapshot { tick: t, .. } = msg {
                assert!(t >= tick);
            }
        }
    }
}
```

---

## 依赖项

### duel-core/Cargo.toml

```toml
[dependencies]
serde = { version = "1", features = ["derive"] }
serde_json = "1"
```

### duel-server/Cargo.toml

```toml
[dependencies]
duel-core = { path = "../duel-core" }
tokio = { version = "1", features = ["full"] }
tokio-tungstenite = "0.21"
futures-util = "0.3"
serde = { version = "1", features = ["derive"] }
serde_json = "1"
clap = { version = "4", features = ["derive"] }
tracing = "0.1"
tracing-subscriber = "0.3"
```

---

## 实施顺序

1. 建workspace + duel-core crate（先只放types + rng）
2. 把TS `createSeededRng`复刻到Rust，写rng测试
3. 移植`DuelBackend`状态机与player/arrow逻辑
4. 实现粒子系统（回归payload需要particles.length）
5. Rust回归测试对齐`regression-baseline.json`
6. 建duel-server crate（WS + tick loop）
7. 写WS冒烟测试
8. CLI server可用

---

## 后续扩展

完成CLI server后：

1. 前端NetClient/NetDriver接入
2. Tauri内置房主开服（复用duel-server代码）
3. 公网房间服务
4. 账号系统接入
