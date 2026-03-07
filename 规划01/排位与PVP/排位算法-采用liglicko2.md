# 排位算法：采用 liglicko2（Glicko-2 即时更新 + 连续时间衰减）

## 选择理由

- 1v1 适配：直接以胜负（Score 0/1）更新双方 rating
- 即时更新：每局结束即可用于下一局匹配（不需要离散 rating period 批处理）
- 不确定性：rating deviation（RD）作为能力范围/置信区间，可用于新号/回流的快速收敛
- 时间衰减：不活跃时 RD 增长，表达“久未对战的不确定性回升”
- 数值稳健：默认参数下提供收敛与非 NaN 的稳定性承诺，并对关键分量做 clamp/限幅

## 工程落点

- third_party：`duel-app/src-tauri/crates/third_party/liglicko2/`
- 门面封装：`duel-app/src-tauri/crates/duel-ranked/`
  - 统一时间换算：`DEFAULT_PERIOD_MS = 86_400_000`（以“天”为 rating period 单位）
  - 统一创建新玩家 rating 的时间锚点：新玩家创建时把 `at` 设为当前 `Instant`

## 输出与用途

- 胜率点估计：`expected_score(first, second, now)` 输出 0..1
- 胜率区间（近似）：`win_prob_band(...)` 以 `rating ± 2*RD` 构造 low/mid/high
- 结算更新：`update_ratings(first, second, score, now)` 输出双方新 rating（包含新 RD 与 σ）

## 与“死亡时间分布”的关系

- liglicko2 仅建模“谁赢/胜率 + 不确定性”，不建模“多久结束”
- 死亡时间（end_tick、duration_ticks）作为独立的对局层统计量记录在服务端，后续可用生存分析/竞争风险模型构造：
  - `p(T_end, winner | rating_gap)` 或 `p(T_end | winner, rating_gap)`

## 验收

- Rust：`cd duel-app/src-tauri && cargo test -p duel-ranked -p duel-server`
- 前端单测：`pnpm -C duel-app test:unit`
- 手测：跑一局 1v1，对局进入结算时终端出现 `ranked_match_recorded` 日志（含 winner、end_tick、duration_ticks、rating_before/after）
