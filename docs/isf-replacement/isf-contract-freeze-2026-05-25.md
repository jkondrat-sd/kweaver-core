# ISF 替换 —— 契约冻结 Spec（Phase 0）

> 日期：2026-05-25　分支：`feat/isf-replacement`
> 目的：钉死「新 auth-service / hydra 必须满足的对外契约」，作为 contract test + 上线影子比对的基准。**不被信创决策（§11.1）阻塞，先行。**
> 上游：`reports/isf-replacement-landing-design-2026-05-25.md`

---

## 1. token introspect 契约（hydra `/admin/oauth2/introspect`）⚠️ 最硬

lib `kweaver-go-lib/hydra.Introspect` 打 ORY 标准 `POST {hydraAdmin}/admin/oauth2/introspect`（form `token=<t>`），按下表解析响应。**新 hydra（含其 login/consent provider = auth-service）必须让 introspect 返回这些字段，否则 lib 类型断言 panic（无 nil 检查）。**

| lib 字段 | 来源 claim | 规则 |
|---|---|---|
| `Active` | `active` (bool) | false 直接返回 |
| `VisitorID` | `sub` (string) | |
| `Scope` | `scope` (string) | |
| `ClientID` | `client_id` (string) | |
| `VisitorTyp` | sub==client_id → `app`；否则 `ext.visitor_type` | 取值 `realname`/`user`/`anonymous`/`app` |
| `LoginIP` | `ext.login_ip` (string) | **仅 VisitorTyp=user 时读** |
| `Udid` | `ext.udid` (string) | 仅 user |
| `AccountTyp` | `ext.account_type` (string) | 仅 user，取值 `other`/`id_card` |
| `ClientTyp` | `ext.client_type` (string) | 仅 user，取值见 ClientType 枚举（windows/ios/android/harmony/mac_os/web/...） |

**硬约束（DoD）**：
- [ ] auth-service 作为 consent provider，session 注入 `ext = {visitor_type, login_ip, udid, account_type, client_type}`，5 个字段对 user 类型必须齐全（否则 lib panic）。
- [ ] app 类型（client_credentials）：sub==client_id，`ext.*` 可缺（lib 走 app 分支不读 ext）。
- [ ] anonymous：VisitorTyp=anonymous，ClientTyp 默认 web。
- [ ] 契约 test：构造 user/app/anonymous 三类 token，断言 introspect 响应可被 lib 正确解析为 `TokenIntrospectInfo`。

> 注：DA 用 `rest.Hydra`（早期 lib 版本，Hydra 类型在 rest 包），与 adp 的 `hydra` 包**异源**。两套 introspect 客户端都打同一 hydra，需**分别**跑契约 test。

---

## 2. authorization 契约（`/api/authorization/v1/*`）

实际只用 RBAC 子集（Deny/Condition/ExpiresAt/obligation 全空，见落地设计 §4）。

| 端点 | 方法 | 请求（实测字段） | 响应 | Casbin 实现 |
|---|---|---|---|---|
| `/operation-check` | POST | `{accessor:{type,id}, resource:{type,id}, operation:[...], method}` | `{result: bool}` | `Enforce(id, "type:id", op)` |
| `/policy` | POST | `[{accessor, resource, operation:{allow:[{id}],deny:[]}, condition:"", expires_at:""}]` | 2xx | `AddPolicy` / `AddGroupingPolicy` |
| `/policy/`（DELETE，Pattern A） + `/policy-delete`（POST，exec-factory）| 两形态 | `[{resource:{type,id}}]` | 2xx | `RemoveFilteredPolicy` |
| `/resource-filter` | POST | `{accessor, resources:[...], operation:[...], allow_operation:bool}` | `{<id>: {id, operation:[...]}}` | 遍历 + `Enforce` |
| `/resource-operation` | POST | 同上 | 同上 | `GetImplicitPermissionsForUser` |
| `/resource-list` | POST | `{accessor, ...}` | 资源列表 | 列 user 可访问 obj |
| `/resource_type/` | — | 资源类型登记（少量） | — | 静态 |

**DoD**：
- [ ] 抓现 ISF 对这 7 端点的真实 req/resp 报文，存为 golden。
- [ ] `policy-delete`（exec-factory）与 `DELETE /policy/`（Pattern A）**双形态都实现**。
- [ ] Casbin model（见 §4）对每端点行为等价，golden 比对一致。

---

## 3. user-management 契约（13 端点，目录查询）

```
/v1/users/  /v1/apps  /v1/apps/  /v1/names  /v2/names  /v1/emails
/v1/departments[/]  /v1/internal-groups[/]  /v1/internal-group-members/
/v1/group-members  /v1/search-org
```
**DoD**：
- [ ] 抓每端点真实 req/resp（TODO：字段级 schema 待从实流量/ISF UserManagement 源码补全）。
- [ ] 标注哪些是读、哪些是写（替换优先实现读）。

---

## 4. Casbin model（验证等价 RBAC 子集）

```ini
[request_definition]
r = sub, obj, act              # sub=accessorID, obj="type:id", act=operation
[policy_definition]
p = sub, obj, act
[role_definition]
g = _, _                       # user/app → role(UUID)
[policy_effect]
e = some(where (p.eft == allow))
[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
```
映射：建对象授权=`AddPolicy(creator,"pipeline:<id>",op...)`；角色授权（DA app_admin）=`AddPolicy("1572fb82-...","agent:*","use")`+`g(user,"1572fb82-...")`；`RESOURCE_ID_ALL "*"`→`keyMatch2`。

**DoD**：
- [ ] 用 §2 golden 报文驱动 Casbin，断言 operation-check/filter/list 结果与 ISF 一致。
- [ ] 确认无需 deny/condition/obligation（已证 kweaver 不用）。

---

## 5. 角色 UUID（保号清单）

| UUID | 角色 | source |
|---|---|---|
| 7dcfcc9c-ad02-11e8-... | 超级管理员 | system |
| d2bd2082-ad03-11e8-... | 系统管理员 | system |
| d8998f72-ad03-11e8-... | 安全管理员 | system |
| def246f2-ad03-11e8-... | 审计管理员 | system |
| e63e1c88-ad03-11e8-... | 组织管理员 | system |
| f06ac18e-ad03-11e8-... | 组织审计员 | system |
| 00990824-4bf7-11f0-... | 数据管理员 | business |
| 3fb94948-5169-11f0-... | AI管理员 | business |
| **1572fb82-526f-11f0-bde6-e674ec8dde71** | **应用管理员** | business（DA `inner_role.go` 硬编码） |

源：`/Users/cx/Work/kweaver-ai/isf/Authorization/driveradapters/init_data/role.json`。**DoD**：新 auth-service seed 沿用同 UUID。

---

## 6. 审计契约（MQ 解耦，低影响）

应用经 `kweaver-go-lib/audit` 发 Kafka topic `AUDIT_TOPIC`（`AuditLog` 结构）。**替换 = 换消费者，应用零改**。`AuditOperator` 由 `audit.TransforOperator(hydra.Visitor)` 转换 → 依赖 §1 的 Visitor 字段。

**DoD**：[ ] 新栈提供 `AUDIT_TOPIC` 消费者（或保留 audit-log）。

---

## 7. 信创/国密（挂 §11.1 决策）
- token/密码加密若需国密 → 用 `kweaver-go-lib/crypto/haitai`（海泰 HSM，SM 算法）。
- hydra DB：信创合规口径决定上游 v2.2 直用 vs rebase fork。**本 spec 不阻塞，待决策。**

---

## 8. Phase 0 待办
- [ ] §1 introspect ext claim 契约 test（user/app/anonymous）
- [ ] §2 抓 authorization 7 端点 golden 报文
- [ ] §3 补 user-management 字段级 schema
- [ ] §4 Casbin model 跑 golden 验等价
- [ ] §5 角色 UUID seed
- [ ] 升级 §11.1 信创合规决策（并行）
