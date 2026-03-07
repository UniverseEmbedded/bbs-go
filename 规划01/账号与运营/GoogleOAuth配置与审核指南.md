# Google OAuth 配置与审核指南

本文档详细说明 Google OAuth 的配置步骤、审核要求以及常见问题解答。

---

## 1. 基础概念

### 1.1 什么是 Google OAuth

Google OAuth 2.0 是一种授权机制：你的应用代表用户调用 Google API（如读取 Drive、Gmail、Calendar 等）时，需要用户在 Google 的授权同意页上点击“允许”，Google 会发放访问令牌。

### 1.2 Scopes 分类

| 类别 | 说明 | 审核要求 |
|------|------|----------|
| non-sensitive | 非敏感权限，如基础身份信息 | 一般不需要审核 |
| sensitive | 敏感权限，如 Gmail、Drive 深度访问 | 需要审核 |
| restricted | 受限权限，如更高级别的数据访问 | 需要审核 + 安全评估 |

> Google Cloud Console 会自动标注你添加的 scopes 属于哪一类。

---

## 2. 审核要求

### 2.1 何时不需要审核

以下情况可以先开发使用，不强制要求审核：

- **开发/测试阶段**：Demo、联调、内测用途
- **个人/小范围自用**：少于 100 用户
- **仅内部使用**：同一个 Google Workspace / Cloud Identity 组织内，设为 internal-only
- **只用 non-sensitive scopes**：只申请非敏感权限时

### 2.2 何时必须审核

满足以下任一条件需要走审核流程：

1. 应用给**任何 Google 账号的外部用户**使用（生产环境），并且
2. 申请的 scopes 包含 **sensitive** 或 **restricted**

### 2.3 不送审的后果

- **未验证应用警告**：用户可能看到 unverified app 警告页
- **用户规模限制**：未验证应用可能被限制在 100 个新用户
- **品牌展示受限**：无法在同意页显示应用名称/Logo

### 2.4 品牌验证（Brand Verification）

如果希望在同意页展示应用名称/Logo，即使只用 non-sensitive scopes，也可能需要做品牌验证。品牌验证要求：
- 在 Google Search Console 验证域名所有权
- 准备隐私政策 URL

---

## 3. 你的场景：外部用户 + 无敏感权限

### 3.1 是否需要送审

| 条件 | 结论 |
|------|------|
| 所有 scopes 都是 non-sensitive | 一般不需要做 OAuth App Verification |
| 想要在同意页展示品牌信息 | 可能需要做 Brand Verification |

### 3.2 关键配置要点：Publishing Status

| 状态 | 限制 |
|------|------|
| Testing | 最多 100 个测试用户可用 |
| In production | 可面向外部用户大规模使用 |

> **重要**：如果停留在 Testing 状态，即使不做审核，外部用户也无法正常使用。

### 3.3 不送审能获得什么权限

只能获得你申请的 **non-sensitive scopes** 对应的能力。敏感/受限范围的权限需要额外审核。

---

## 4. 配置步骤

### 4.1 Google Developer Console 配置

1. 创建 OAuth 2.0 客户端 ID
2. 配置回调 URL：`https://pama1234.tech/api/auth/oauth/callback/google`
3. 获取客户端 ID 和客户端密钥

OAuth 路由规范以 [账号系统设计.md](./账号系统设计.md) 为准。

### 4.2 OAuth 同意屏幕配置

路径：**Google Cloud Console → APIs & Services → OAuth consent screen**

需要填写：
- App name（应用名称）
- User support email（支持邮箱）
- Developer contact info（开发者联系方式）
- Authorized domains（授权域名，需先在 Search Console 验证）
- Privacy policy URL（隐私政策，必填）
- Scopes（只选你真正需要的，越少越好）

### 4.3 Publishing Status 设置

- 初期可设为 Testing 用于内部测试
- 正式对外发布前改为 In production

---

## 5. 审核流程（如果需要）

### 5.1 准备材料

- 隐私政策 URL（必填）
- 应用主页/官网 URL（建议）
- 授权域名所有权验证（Search Console）
- 演示视频（申请 sensitive/restricted 时需要）
- 测试账号（审核人员可能需要）

### 5.2 常见被退回原因

- 域名未在 Search Console 验证
- 隐私政策不合格（缺少数据收集/用途/共享/删除说明）
- 首页缺少指向隐私政策/服务条款的可见链接（仅在 about 页提供链接可能不被认可）
- OAuth 同意屏幕 App name 与首页可见应用名称不一致
- 演示视频未覆盖每个敏感 scope 的用户旅程
- scope 申请过大（只需要只读却申请读写）

### 5.3 品牌审核修复要点（不涉及敏感信息）

- 首页所有权：用 Search Console 验证域名所有权后，确保同意屏幕的 Authorized domains 仅填写域名（不含协议与路径）
- 隐私/条款：确保首页（`/`）包含可点击的 Privacy Policy / Terms of Service 链接，且页面对公网 200 可访问
- 应用名称一致：同意屏幕 App name 必须与首页可见应用名称一致（建议首页显式展示应用名，而不只依赖浏览器标题）
- 记录原则：文档中只记录流程、验收点、常见坑；不写入 OAuth client secret、cookie、审核账号等敏感信息

---

## 6. 参考链接

| 资源 | 链接 |
|------|------|
| OAuth App Verification Help | https://support.google.com/cloud/answer/13463073 |
| Unverified apps 说明 | https://support.google.com/googleapi/answer/7454865 |
| Brand verification 指南 | https://developers.google.com/identity/protocols/oauth2/production-readiness/brand-verification |
| 配置 OAuth 同意屏幕 | https://developers.google.com/workspace/guides/configure-oauth-consent |

---

## 相关文档

- [账号系统设计.md](./账号系统设计.md) - 账号系统整体设计
- [Go服务端集成方案.md](./Go服务端集成方案.md) - 服务端 OAuth 配置

---

## 7. 部署记录

### 2026-03-04

已部署隐私政策和服务条款页面（英文），用于满足 Google OAuth 审核要求：

| 文件 | 说明 |
|------|------|
| `server/web_page/privacy.html` | 隐私政策页面 |
| `server/web_page/terms.html` | 服务条款页面 |
| `server/web_page/about.html` | 在关于页面底部添加了 Privacy Policy / Terms of Service 链接 |

**页面 URL**：
- Privacy Policy: https://pama1234.tech/privacy
- Terms of Service: https://pama1234.tech/terms

**用于 Google OAuth 审核**：
- OAuth Consent Screen 已填写：
  - Privacy Policy URL: https://pama1234.tech/privacy
  - Terms of Service URL: https://pama1234.tech/terms
- 域名 pama1234.tech 已加入 Authorized domains
