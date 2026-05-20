# VEGA Part 9: Resource 行列规则权限设计

---

## 一、需求分析

### 1.1 需求背景

数据资源需要具备细粒度的数据访问控制能力，保证数据安全。在 VEGA 资源管理层级架构中，Resource 作为具体的数据资产实体，需要支持行列级别的权限管控。

### 1.2 设计目标

1. **功能实现**：设计一套标准化接口，覆盖上层业务对于组织和人数据权限的使用。
2. **功能优化**：对现有授权功能进行梳理，识别并剥离不合理的功能，精简设计复杂度，提升可维护性。
3. **架构统一**：将行列规则权限能力从特定视图类型扩展到所有 Resource 类型。

### 1.3 需求价值

1. **降低维护成本**：通过剥离冗余功能，简化系统架构，减少潜在的维护问题和故障点，提升系统稳定性。
2. **数据安全管控**：构建完善的数据访问控制能力，保障核心数据资产安全。
3. **统一权限模型**：为所有 Resource 类型提供一致的行列规则权限管理能力。

### 1.4 业务目标

1. 实现对 Resource 行列级权限的支持。
2. 基于统一的权限框架建立统一的用户和部门操作鉴权体系。

---

## 二、设计方案

### 2.1 授权概念定义

| 概念 | 描述 |
|------|------|
| **授权** | 能够授权给别人行列规则、查看行列规则、新建、编辑、删除已有的行列规则。 |
| **授权仅分配** | 能够授权给别人行列规则、查看行列规则、不能新建、编辑、删除已有的行列规则。 |

### 2.2 权限点设计

Resource 权限点列表（加粗为新增）：
- 新建、编辑、删除、查看、权限管理、数据查询、导入、导出、**行列规则管理**、**行列规则授权**

### 2.3 规则应用范围

创建行列规则时授权范围可选择 Resource，支持复制已有行列规则模板功能。

---

## 三、资源类型定义

### 3.1 权限码

#### 3.1.1 Resource 权限码

| 序号 | 操作 | 编码（唯一标识） | 说明 |
|------|------|------------------|------|
| 1 | 查看 | view_detail | 获取资源的信息、具体的配置等 |
| 2 | 新建 | create | 生成一个新的资源实例 |
| 3 | 编辑 | modify | 更新已有资源的内容或属性 |
| 4 | 删除 | delete | 删除已存在的资源 |
| 5 | 数据查询 | data_query | 数据资源对象的数据内容查询 |
| 6 | 权限管理 | authorize | 赋予用户或角色对资源的操作权限 |
| 7 | 导入 | import | 导入资源 |
| 8 | 导出 | export | 导出资源 |
| 9 | **行列规则管理** | **rule_manage** | **对资源行列规则增删改查** |
| 10 | **行列规则授权** | **rule_authorize** | **赋予用户或角色对资源行列规则的操作权限** |

### 3.2 资源类型列表

| 菜单 | 模块 | 资源类型中文名 | 资源类型 id | 备注 |
|------|------|----------------|------------|------|
| 数据资源 | 数据资源 | 数据资源 | resource | 涵盖 table、index、file、fileset、logicview 等类别 |
| -- | -- | 资源行列规则 | resource_row_column_rule | 资源行列规则在单独的授权页面，不会在信息安全管理的菜单显示 |

### 3.3 资源类型初始化

#### 3.3.1 Resource 资源初始化

```json
{
    "id": "resource",
    "name": "数据资源",
    "description": "数据资源的配置信息",
    "instance_url": "GET /api/vega-backend/v1/resources",
    "data_struct": "string",
    "operation": [
        {
            "id": "view_detail",
            "name": [
                {"language": "zh_cn", "value": "查看"},
                {"language": "en_us", "value": "view_detail"},
                {"language": "zh_tw", "value": "查看"}
            ],
            "description": "查看数据资源",
            "scope": ["type", "instance"]
        },
        {
            "id": "create",
            "name": [
                {"language": "zh_cn", "value": "新建"},
                {"language": "en_us", "value": "create"},
                {"language": "zh_tw", "value": "新建"}
            ],
            "description": "新建数据资源",
            "scope": ["type", "instance"]
        },
        {
            "id": "modify",
            "name": [
                {"language": "zh_cn", "value": "编辑"},
                {"language": "en_us", "value": "modify"},
                {"language": "zh_tw", "value": "編輯"}
            ],
            "description": "编辑数据资源",
            "scope": ["type", "instance"]
        },
        {
            "id": "delete",
            "name": [
                {"language": "zh_cn", "value": "删除"},
                {"language": "en_us", "value": "delete"},
                {"language": "zh_tw", "value": "刪除"}
            ],
            "description": "删除数据资源",
            "scope": ["type", "instance"]
        },
        {
            "id": "data_query",
            "name": [
                {"language": "zh_cn", "value": "数据查询"},
                {"language": "en_us", "value": "data_query"},
                {"language": "zh_tw", "value": "數據查詢"}
            ],
            "description": "数据资源的数据内容查询",
            "scope": ["type", "instance"]
        },
        {
            "id": "authorize",
            "name": [
                {"language": "zh_cn", "value": "权限管理"},
                {"language": "en_us", "value": "authorize"},
                {"language": "zh_tw", "value": "權限管理"}
            ],
            "description": "赋予用户或角色对资源的操作权限",
            "scope": ["type", "instance"]
        },
        {
            "id": "import",
            "name": [
                {"language": "zh_cn", "value": "导入"},
                {"language": "en_us", "value": "import"},
                {"language": "zh_tw", "value": "導入"}
            ],
            "description": "导入资源",
            "scope": ["type", "instance"]
        },
        {
            "id": "export",
            "name": [
                {"language": "zh_cn", "value": "导出"},
                {"language": "en_us", "value": "export"},
                {"language": "zh_tw", "value": "導出"}
            ],
            "description": "导出资源",
            "scope": ["type", "instance"]
        },
        {
            "id": "rule_manage",
            "name": [
                {"language": "zh_cn", "value": "行列规则管理"},
                {"language": "en_us", "value": "rule_manage"},
                {"language": "zh_tw", "value": "行列規則管理"}
            ],
            "description": "对资源行列规则增删改查",
            "scope": ["type", "instance"]
        },
        {
            "id": "rule_authorize",
            "name": [
                {"language": "zh_cn", "value": "行列规则授权"},
                {"language": "en_us", "value": "rule_authorize"},
                {"language": "zh_tw", "value": "行列規則授權"}
            ],
            "description": "赋予用户或角色对资源行列规则的操作权限",
            "scope": ["type", "instance"]
        }
    ]
}
```

---

## 四、权限策略示例

### 4.1 策略结构

```json
{
    "accessor": {
        "type": "user",
        "id": "user-uuid"
    },
    "resource": {
        "type": "resource",
        "id": "resource-uuid",
        "name": "用户信息表"
    },
    "operation": {
        "allow": [
            {"id": "view_detail"},
            {"id": "data_query"}
        ],
        "deny": []
    },
    "condition": "",
    "expires_at": "1970-01-01T08:00:00+08:00"
}
```

### 4.2 示例策略

```json
[
    {
        "expires_at": "1970-01-01T08:00:00+08:00",
        "resource": {
            "id": "*",
            "type": "resource",
            "name": "数据资源"
        },
        "accessor": {
            "id": "3fb94948-5169-11f0-b662-3a7bdba2913f",
            "type": "role",
            "name": "AI 管理员"
        },
        "operation": {
            "allow": [
                {"id": "view_detail"},
                {"id": "create"},
                {"id": "data_query"},
                {"id": "import"},
                {"id": "export"}
            ],
            "deny": []
        }
    }
]
```

---

## 五、功能逻辑设计

### 5.1 创建资源行列规则参数

| 字段 | 类型 | 描述 | 备注 | 格式示例 |
|------|------|------|------|----------|
| id | string | 行列规则 id | 支持自定义，若未自定义系统自动生成；新建后不可更改；只能包含小写英文字母、数字、下划线、连字符，且不能以下划线和连字符开头 | |
| name | string | 行列规则名称 | 唯一；长度不超过 255 字符 | |
| resource_id | string | 资源 id | 必填 | |
| tags | []string | 标签 | | |
| description | string | 备注 | | |
| fields | []string | 列 | 字段 name 列表 | |
| row_filters | condCfg | 行过滤规则 | 支持单层及多层嵌套 | 见下文示例 |

#### 5.1.1 行过滤规则结构体定义

行过滤规则使用 `FilterCondCfg` 结构体，支持条件表达式与嵌套组合：

```go
type ValueOptCfg struct {
    ValueFrom string `json:"value_from,omitempty"` // 值来源: const/field/user
    Value     any    `json:"value,omitempty"`       // 值或引用标识
    RealValue any    `json:"real_value,omitempty"`  // 运行时解析后的真实值
}

type FilterCondCfg struct {
    Name        string           `json:"field,omitempty"`         // 字段名（叶子节点）
    Operation   string           `json:"operation,omitempty"`     // 操作符: ==, !=, >, >=, <, <=, in, like, and, or, match, exists 等
    SubConds    []*FilterCondCfg `json:"sub_conditions,omitempty"` // 子条件列表（and/or 复合节点）
    ValueOptCfg `mapstructure:",squash"`                           // 值配置（嵌入）
    RemainCfg   map[string]any  `mapstructure:",remain"`           // 扩展配置
}
```

**支持的 value_from 类型**：

| value_from | 说明 | value 示例 |
|-----------|------|-----------|
| `const` | 常量值 | `"group"`, `"active"`, `100` |
| `field` | 引用另一个字段的值 | `"status"` |
| `user` | 引用当前用户属性 | `"department_id"`, `"user_id"` |

**行过滤规则格式示例**：

**单层级（叶子条件）**：
```json
{
    "operation": "==",
    "field": "f4",
    "value_from": "const",
    "value": "group"
}
```

**多层级 and/or 嵌套（复合条件）**：
```json
{
    "operation": "and",
    "sub_conditions": [
        {
            "operation": "or",
            "sub_conditions": [
                {"operation": "==", "field": "f1", "value_from": "const", "value": "a"},
                {"operation": "==", "field": "f1", "value_from": "const", "value": "b"}
            ]
        },
        {
            "operation": "==",
            "field": "f2",
            "value_from": "const",
            "value": "c"
        }
    ]
}
```

### 5.2 更新资源行列规则参数

与创建参数一致，不需要传入 `id`、`resource_id`、`creator`、`create_time`，这些字段由系统维护。**resource_id 不允许变更**。

### 5.3 删除资源行列规则

| 字段 | 类型 | 描述 | 备注 |
|------|------|------|------|
| rule_ids | string | 规则 ID 列表 | 必填，通过 URL 路径参数传递，多个 ID 逗号分隔 |

### 5.4 查询资源行列规则

| 字段 | 类型 | 描述 | 备注 |
|------|------|------|------|
| resource_id | string | 资源 ID | 可选 |
| name | string | 规则名称 | 可选，精确匹配 |
| name_pattern | string | 规则名称模糊 | 可选，支持模糊搜索 |
| tag | string | 标签 | 可选 |
| offset | int | 偏移量 | 默认 0 |
| limit | int | 每页数量 | 默认 20，最大 1000 |
| sort | string | 排序字段 | 可选：name / create_time / update_time |
| direction | string | 排序方向 | 可选：asc / desc，默认 desc |

### 5.5 参数校验规则

#### 5.5.1 创建校验

| 字段 | 校验规则 | 错误码 |
|------|---------|--------|
| name | 必填，长度 ≤ 255 字符 | VegaBackend.InvalidParameter.RuleName |
| resource_id | 必填 | VegaBackend.InvalidParameter.ResourceID |
| id | 可选；若提供必须匹配正则 `^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$` | PublicError.BadRequest |
| tags | 每个 tag 长度 ≤ 255 字符，不含 ` ,` 字符，总数 ≤ 10 | VegaBackend.InvalidParameter.Tags |

#### 5.5.2 更新校验

| 字段 | 校验规则 | 错误码 |
|------|---------|--------|
| name | 可选；若提供则长度 ≤ 255 字符 | VegaBackend.InvalidParameter.RuleName |
| resource_id | 不允许变更，新旧值必须一致 | PublicError.BadRequest |
| tags | 同创建校验 | VegaBackend.InvalidParameter.Tags |

### 5.6 级联删除

当 Resource（数据资源）被删除时，系统自动级联删除关联的所有行列规则及其权限策略。级联流程：

1. 通过 `GetSimpleRulesByResourceIDs` 获取资源下所有规则 ID
2. 调用 `DeleteResourceRowColumnRules` 删除数据库记录
3. 调用 `ps.DeleteResources` 清理权限策略

此操作在 Service 层通过 `DeleteRowColumnRulesByResourceIDs` 方法暴露，内部使用，不校验当前用户权限（由调用方保证）。

---

## 六、数据库设计

### 6.1 表结构

**表名**：`t_resource_row_column_rule`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| f_rule_id | VARCHAR(64) | PRIMARY KEY | 规则 ID |
| f_rule_name | VARCHAR(255) | NOT NULL, INDEX | 规则名称 |
| f_resource_id | VARCHAR(64) | NOT NULL, INDEX | 资源 ID |
| f_tags | VARCHAR(1024) | DEFAULT '' | 标签（逗号分隔） |
| f_comment | VARCHAR(1024) | DEFAULT '' | 备注 |
| f_fields | TEXT | | 字段列表（JSON 数组） |
| f_row_filters | TEXT | | 行过滤规则（JSON） |
| f_create_time | BIGINT | NOT NULL | 创建时间（毫秒时间戳） |
| f_update_time | BIGINT | NOT NULL | 更新时间（毫秒时间戳） |
| f_creator | VARCHAR(64) | NOT NULL | 创建者 ID |
| f_creator_type | VARCHAR(32) | NOT NULL | 创建者类型（user/app/role） |
| f_updater | VARCHAR(64) | NOT NULL | 更新者 ID |
| f_updater_type | VARCHAR(32) | NOT NULL | 更新者类型（user/app/role） |

### 6.2 索引设计

- PRIMARY KEY: f_rule_id
- INDEX idx_resource_id: f_resource_id
- INDEX idx_rule_name: f_rule_name
- INDEX idx_update_time: f_update_time

---

## 七、API 接口设计

### 7.1 创建资源行列规则

**接口**：`POST /api/vega-backend/v1/resource-row-column-rules`

**请求体**：
```json
{
    "id": "rule-001",
    "name": "部门数据隔离规则",
    "resource_id": "resource-001",
    "tags": ["security", "department"],
    "comment": "按部门隔离数据",
    "fields": ["f1", "f2", "f3"],
    "row_filters": {
        "operation": "==",
        "field": "department",
        "value_from": "user",
        "value": "department_id"
    }
}
```

**响应**：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "rule_ids": ["rule-001"]
    }
}
```

### 7.2 更新资源行列规则

**接口**：`PUT /api/vega-backend/v1/resource-row-column-rules/{rule_id}`

**请求体**：
```json
{
    "name": "更新后的规则名称",
    "tags": ["security"],
    "comment": "更新备注",
    "fields": ["f1", "f2"],
    "row_filters": {...}
}
```

### 7.3 删除资源行列规则

**接口**：`DELETE /api/vega-backend/v1/resource-row-column-rules/{rule_ids}`

**路径参数**：
- `rule_ids`: 规则 ID 列表，多个 ID 用逗号分隔（例如：`rule-001,rule-002,rule-003`）

**响应**：
```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

### 7.4 查询单个资源行列规则

**接口**：`GET /api/vega-backend/v1/resource-row-column-rules/{rule_id}`

**响应**：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "rule": {
            "id": "rule-001",
            "name": "部门数据隔离规则",
            "resource_id": "resource-001",
            "resource_name": "用户信息表",
            "tags": ["security", "department"],
            "comment": "按部门隔离数据",
            "fields": ["f1", "f2", "f3"],
            "row_filters": {...},
            "create_time": 1234567890000,
            "update_time": 1234567890000,
            "creator": {
                "id": "user-001",
                "type": "user"
            },
            "updater": {
                "id": "user-001",
                "type": "user"
            },
            "operations": ["rule_manage", "rule_authorize"]
        }
    }
}
```

### 7.5 列表查询资源行列规则

**接口**：`GET /api/vega-backend/v1/resource-row-column-rules`

**查询参数**：
| 参数 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| resource_id | string | 资源 ID 过滤 | - |
| name | string | 规则名称精确匹配 | - |
| name_pattern | string | 规则名称模糊匹配 | - |
| tag | string | 标签过滤 | - |
| offset | int | 偏移量 | 0 |
| limit | int | 每页数量（最大 1000） | 20 |
| sort | string | 排序字段（name/create_time/update_time） | update_time |
| direction | string | 排序方向（asc/desc） | desc |

**响应**：
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "rules": [...]
    }
}
```

### 7.6 业务响应说明

所有接口统一使用 `rest.ReplyOK` / `rest.ReplyError` 进行响应封装：

- 成功：`HTTP 200`，Body 为 `{"code": 0, "message": "success", "data": {...}}`
- 业务错误：对应 HTTP 状态码，Body 为 `{"code": "<错误码>", "message": "<错误描述>", "error_details": "<详细信息>", "request_id": "<追踪ID>"}`

---

## 八、路由设置

### 8.1 外部 API 路由（External）

外部 API 需要通过 OAuth2 认证，所有请求必须携带有效的访问令牌。

| 方法 | 路由路径 | 说明 | 权限要求 |
|------|---------|------|---------|
| POST | `/api/vega-backend/v1/resource-row-column-rules` | 创建资源行列规则 | `rule_manage` |
| GET | `/api/vega-backend/v1/resource-row-column-rules` | 列表查询资源行列规则 | `rule_manage` 或 `rule_authorize` |
| GET | `/api/vega-backend/v1/resource-row-column-rules/:rule_id` | 查询单个资源行列规则 | `rule_manage` 或 `rule_authorize` |
| PUT | `/api/vega-backend/v1/resource-row-column-rules/:rule_id` | 更新资源行列规则 | `rule_manage` |
| DELETE | `/api/vega-backend/v1/resource-row-column-rules/:rule_ids` | 删除资源行列规则 | `rule_manage` |

**路由注册代码**（router.go）:
```go
// Resource Row Column Rule APIs - External
resourceRowColumnRules := apiV1.Group("/resource-row-column-rules")
{
    resourceRowColumnRules.POST("", r.verifyJsonContentType(), r.CreateResourceRowColumnRulesByEx)
    resourceRowColumnRules.GET("", r.ListResourceRowColumnRulesByEx)
    resourceRowColumnRules.GET("/:rule_id", r.GetResourceRowColumnRulesByEx)
    resourceRowColumnRules.PUT("/:rule_id", r.verifyJsonContentType(), r.UpdateResourceRowColumnRulesByEx)
    resourceRowColumnRules.DELETE("/:rule_ids", r.DeleteResourceRowColumnRulesByEx)
}
```

### 8.2 内部 API 路由（Internal）

内部 API 用于系统内部服务间调用，通过请求头传递用户身份信息，不需要 OAuth2 认证。

| 方法 | 路由路径 | 说明 | 权限要求 |
|------|---------|------|---------|
| POST | `/api/vega-backend/in/v1/resource-row-column-rules` | 创建资源行列规则 | `rule_manage` |
| GET | `/api/vega-backend/in/v1/resource-row-column-rules` | 列表查询资源行列规则 | `rule_manage` 或 `rule_authorize` |
| GET | `/api/vega-backend/in/v1/resource-row-column-rules/:rule_id` | 查询单个资源行列规则 | `rule_manage` 或 `rule_authorize` |
| PUT | `/api/vega-backend/in/v1/resource-row-column-rules/:rule_id` | 更新资源行列规则 | `rule_manage` |
| DELETE | `/api/vega-backend/in/v1/resource-row-column-rules/:rule_ids` | 删除资源行列规则 | `rule_manage` |

**路由注册代码**（router.go）:
```go
// Resource Row Column Rule APIs - Internal
resourceRowColumnRules := apiInV1.Group("/resource-row-column-rules")
{
    resourceRowColumnRules.POST("", r.verifyJsonContentType(), r.CreateResourceRowColumnRulesByIn)
    resourceRowColumnRules.GET("", r.ListResourceRowColumnRulesByIn)
    resourceRowColumnRules.GET("/:rule_id", r.GetResourceRowColumnRulesByIn)
    resourceRowColumnRules.PUT("/:rule_id", r.verifyJsonContentType(), r.UpdateResourceRowColumnRulesByIn)
    resourceRowColumnRules.DELETE("/:rule_ids", r.DeleteResourceRowColumnRulesByIn)
}
```

### 8.3 路由说明

1. **外部 API** (`/api/vega-backend/v1/`):
   - 需要 OAuth2 认证
   - 从 JWT token 中提取用户信息
   - 面向外部客户端

2. **内部 API** (`/api/vega-backend/in/v1/`):
   - 不需要 OAuth2 认证
   - 从请求头 `X-Account-Id` 和 `X-Account-Type` 获取用户信息
   - 面向内部服务间调用

3. **Content-Type 验证**:
   - POST、PUT 请求必须携带 `Content-Type: application/json` 头
   - 通过 `verifyJsonContentType()` 中间件进行验证

4. **路径参数**:
   - `:rule_id`: 单个规则 ID
   - `:rule_ids`: 多个规则 ID，逗号分隔（例如：`rule-001,rule-002,rule-003`）

---

## 九、权限检查流程

### 9.1 操作与权限映射

| 操作 | 接口 | 所需权限 | 说明 |
|------|------|---------|------|
| 创建规则 | POST | `view_detail` + `rule_manage` | 对目标资源必须同时具有查看和规则管理权限 |
| 查询单个规则 | GET /:rule_id | `view_detail` | 对规则本身需要查看权限 |
| 列表查询规则 | GET | `rule_manage` 或 `rule_authorize` | 有任一权限即可看到规则列表 |
| 更新规则 | PUT /:rule_id | `view_detail` + `rule_manage` | 对规则所在资源必须同时具有查看和规则管理权限 |
| 删除规则 | DELETE /:rule_ids | `rule_manage` | 对规则所在资源必须具有规则管理权限 |

### 9.2 权限决策流程

```
┌─────────────────────────────────────────────────┐
│                 客户端请求                         │
├─────────────────────────────────────────────────┤
│  ① Handler 层：解析用户身份（OAuth2 / 内部头）      │
│  ② Handler 层：参数校验                            │
│  ③ Service 层：调用 PermissionService 鉴权         │
│     ├─ FilterResources：过滤有权限的资源列表        │
│     └─ CheckPermission：检查对特定资源的操作权限    │
│  ④ Service 层：业务逻辑处理                        │
│  ⑤ Access 层：数据库操作                           │
│  ⑥ Service 层：审计日志记录                        │
└─────────────────────────────────────────────────┘
```

### 9.3 创建/更新/删除时的权限检查

1. Service 层调用 `ps.FilterResources` 或 `ps.CheckPermission` 进行策略决策
2. FilterResources 返回有权限的资源 ID → 操作集合的映射
3. 遍历请求中的 resource_id，确认每个 ID 都在有权限的映射中
4. 若任一 resource_id 无权限，返回 `403 Forbidden`

### 9.4 列表查询时的权限过滤

列表查询在 Service 层 `ListResourceRowColumnRules` 中完成权限过滤：

1. Access 层根据查询条件获取所有匹配的规则记录（不分页）
2. Service 层调用 `ps.FilterResources` 获取当前用户对每个 resource_id 的操作权限
3. 筛选出具有 `rule_manage` **或** `rule_authorize` 任意一个权限的规则
4. 对通过权限过滤的结果集进行内存分页
5. 返回分页后的结果和过滤后的总数

> 注意：内部请求（`IsInnerRequest=true`）跳过权限过滤，返回原始全部数据。

### 9.5 查询单个规则时的权限检查

1. Access 层根据 rule_id 获取规则信息
2. Service 层调用 `ps.FilterResources` 检查用户是否有对应资源的 `view_detail` 权限
3. 无权限时返回 `403 Forbidden`

---

## 十、架构层次设计

### 10.1 分层架构

```
┌─────────────────────────────────────────────────────┐
│                    Handler 层                        │
│   driveradapters/resource_row_column_rule_handler.go │
│   - HTTP 请求/响应处理                                │
│   - OAuth2 认证 / 内部身份注入                        │
│   - 请求参数绑定与基础校验                             │
│   - 审计日志记录                                      │
├─────────────────────────────────────────────────────┤
│                    Service 层                         │
│   logics/resource/resource_row_column_rule_service.go │
│   - 业务逻辑编排                                      │
│   - 权限校验（调用 PermissionService）                 │
│   - 规则名称唯一性校验                                │
│   - 参数校验                                          │
│   - ID 自动生成                                       │
├─────────────────────────────────────────────────────┤
│                    Access 层                          │
│   drivenadapters/resource/resource_row_column_rule_access.go │
│   - 数据库 CRUD 操作                                   │
│   - SQL 构建（使用 squirrel 库）                       │
│   - 数据序列化/反序列化（JSON 字段）                   │
├─────────────────────────────────────────────────────┤
│                    数据存储                            │
│   t_resource_row_column_rule                          │
└─────────────────────────────────────────────────────┘
```

### 10.2 接口定义

```go
// Access 层接口
type ResourceRowColumnRuleAccess interface {
    CreateResourceRowColumnRules(ctx context.Context, rules []*ResourceRowColumnRule) error
    UpdateResourceRowColumnRule(ctx context.Context, rule *ResourceRowColumnRule) error
    GetResourceRowColumnRules(ctx context.Context, ruleIDs []string) ([]*ResourceRowColumnRule, error)
    ListResourceRowColumnRules(ctx context.Context, query *ListRowColumnRuleQueryParams) ([]*ResourceRowColumnRule, error)
    DeleteResourceRowColumnRules(ctx context.Context, tx *sql.Tx, ruleIDs []string) error
    CheckResourceRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, bool, error)
    CheckResourceRowColumnRuleExistByName(ctx context.Context, ruleName, resourceID string) (string, bool, error)
    GetSimpleRulesByResourceIDs(ctx context.Context, tx *sql.Tx, resourceIDs []string) ([]*ResourceRowColumnRule, error)
    GetSimpleRulesByRuleIDs(ctx context.Context, tx *sql.Tx, ruleIDs []string) ([]*ResourceRowColumnRule, error)
}

// Service 层接口
type ResourceRowColumnRuleService interface {
    CreateResourceRowColumnRules(ctx context.Context, rules []*ResourceRowColumnRule) ([]string, error)
    UpdateResourceRowColumnRule(ctx context.Context, rule *ResourceRowColumnRule) error
    DeleteResourceRowColumnRules(ctx context.Context, ruleIDs []string) error
    DeleteRowColumnRulesByResourceIDs(ctx context.Context, tx *sql.Tx, resourceIDs []string) error
    GetResourceRowColumnRules(ctx context.Context, ruleIDs []string) ([]*ResourceRowColumnRule, error)
    ListResourceRowColumnRules(ctx context.Context, params *ListRowColumnRuleQueryParams) ([]*ResourceRowColumnRule, int, error)
    ListResourceRowColumnRuleSrcs(ctx context.Context, params *ListRowColumnRuleQueryParams) ([]PermissionResource, int, error)
    CheckResourceRowColumnRuleExistByID(ctx context.Context, ruleID string) (string, error)
    CheckResourceRowColumnRuleExistByName(ctx context.Context, ruleName, resourceID string) (string, bool, error)
}
```

### 10.3 核心数据结构

```go
// 资源行列规则
type ResourceRowColumnRule struct {
    RuleID       string         `json:"id"`
    RuleName     string         `json:"name"`
    ResourceID   string         `json:"resource_id"`
    ResourceName string         `json:"resource_name,omitempty"`
    Tags         []string       `json:"tags"`
    Comment      string         `json:"comment"`
    CreateTime   int64          `json:"create_time"`
    UpdateTime   int64          `json:"update_time"`
    Creator      AccountInfo    `json:"creator"`
    Updater      AccountInfo    `json:"updater"`
    Fields       []string       `json:"fields"`
    RowFilters   *FilterCondCfg `json:"row_filters"`
    Operations   []string       `json:"operations,omitempty"`  // 当前用户的可操作权限
}

// 列表查询参数
type ListRowColumnRuleQueryParams struct {
    Name           string
    NamePattern    string
    ResourceID     string
    Tag            string
    IsInnerRequest bool   // 内部请求跳过权限过滤
    PaginationQueryParams
}

// 过滤条件配置
type FilterCondCfg struct {
    Name        string           `json:"field,omitempty"`
    Operation   string           `json:"operation,omitempty"`
    SubConds    []*FilterCondCfg `json:"sub_conditions,omitempty"`
    ValueOptCfg                  // 嵌入值配置
    RemainCfg map[string]any     // 扩展配置
}

type ValueOptCfg struct {
    ValueFrom string `json:"value_from,omitempty"`
    Value     any    `json:"value,omitempty"`
    RealValue any    `json:"real_value,omitempty"`
}

type AccountInfo struct {
    ID   string `json:"id"`
    Type string `json:"type"`    // user / app / role
}
```

---

## 十一、错误码参考

### 11.1 资源行列规则特有错误码

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | VegaBackend.InvalidParameter.RuleName | 规则名称参数无效 |
| 400 | VegaBackend.InvalidParameter.ResourceID | 资源 ID 参数无效 |
| 400 | VegaBackend.InvalidParameter.RuleID | 规则 ID 参数无效 |
| 400 | VegaBackend.InvalidParameter.Tags | 标签参数无效 |
| 400 | VegaBackend.ResourceRowColumnRule.ExistByName | 规则名称已存在 |
| 403 | PublicError.Forbidden | 权限不足 |
| 404 | VegaBackend.ResourceRowColumnRule.NotFound | 规则不存在 |
| 404 | VegaBackend.Resource.NotFound | 资源不存在 |
| 500 | PublicError.InternalServerError | 内部服务错误 |

### 11.2 公共错误码

| HTTP 状态码 | 错误码 | 说明 |
|-------------|--------|------|
| 400 | PublicError.BadRequest | 请求参数错误 |
| 400 | VegaBackend.InvalidParameter.RequestBody | 请求体绑定失败 |

---

## 十二、审计日志

### 12.1 审计事件

| 操作 | audit.OPERATION | 触发点 |
|------|----------------|--------|
| 创建规则 | CREATE | Handler 层 `CreateResourceRowColumnRules` 完成后 |
| 查询规则 | read | Handler 层 `GetResourceRowColumnRules` / `ListResourceRowColumnRules` 完成后 |
| 更新规则 | UPDATE | Handler 层 `UpdateResourceRowColumnRules` 完成后 |
| 删除规则 | DELETE | Handler 层 `DeleteResourceRowColumnRules` 完成后 |

### 12.2 审计日志结构

```go
type AuditObject struct {
    Type string      // 操作对象类型
    ID   string      // 操作对象 ID（规则 ID）
    Name string      // 操作对象名称（规则名称）
}
```

审计日志通过 `audit.NewInfoLog`（成功）或 `audit.NewWarnLogWithError`（失败）记录，包含操作者身份、操作类型、操作对象和错误详情。

---

## 十三、安全考虑

1. **权限验证**：所有操作均通过 PermissionService 进行权限验证，确保只有具备相应操作权限的用户才能执行
2. **规则隔离**：不同资源的规则相互隔离，不可跨资源应用或访问
3. **资源 ID 不变性**：规则一旦创建，关联的 resource_id 不可变更
4. **审计日志**：所有规则的创建、修改、删除操作均记录审计日志
5. **级联清理**：删除资源时自动级联删除关联的行列规则，避免孤儿数据
6. **数据序列化**：fields 和 row_filters 以 JSON 格式持久化，通过 sonic 库进行高效序列化/反序列化

---

## 十四、版本历史

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|----------|------|
| 0.1.0 | 2024-01-01 | 初始版本 | - |
| 0.7.0 | 2024-06-01 | 扩展到所有 Resource 类型 | - |
| 1.0.0 | 2026-05-21 | 根据实现代码完善文档：修正 DELETE API（路径参数），补充架构分层、错误码、审计日志、级联删除、FilterCondCfg 类型定义、详细权限流程、参数校验规则等 | - |
