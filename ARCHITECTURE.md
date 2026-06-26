# 社区宠物诊所后端 — 架构说明

> 目标读者：刚接手项目的新同事，半小时内搞懂系统怎么跑、代码怎么改。

---

## 一、项目一眼看懂

```
project123/
├── main.go                  ← 程序入口，启动顺序：配置→MySQL→Redis→路由→监听端口
├── router.go                ← 所有 API 路由定义 + 中间件挂载
├── config/
│   └── config.go            ← 读 .env / 环境变量，全局单例 AppConfig
├── database/
│   ├── db.go                ← MySQL 连接 + GORM AutoMigrate 自动建表
│   └── redis.go             ← Redis 连接 + 排班缓存读写/失效
├── common/
│   ├── response.go          ← 统一 JSON 响应格式 + 全部错误码定义
│   └── validator.go         ← 通用校验工具函数（CheckExists / CheckOwned / Ensure）
├── middleware/
│   └── jwt.go               ← JWT 签发/解析 + JWTAuth 中间件 + RoleAuth 角色拦截
├── models/
│   └── models.go            ← 5 张表的结构体定义（User / Doctor / Service / Pet / Schedule / Appointment）
└── handlers/
    ├── auth_handler.go      ← 模块1：用户认证（注册/登录/查当前用户）
    ├── schedule_handler.go  ← 模块2：医生排班（CRUD + Redis 缓存）
    ├── appointment_handler.go ← 模块3：预约管理（创建/确认/拒绝/取消 + 冲突检测）
    ├── service_handler.go   ← 模块4：服务项目 + 医生管理（CRUD）
    └── pet_handler.go       ← 宠物档案（CRUD，owner_id 严格隔离）
```

---

## 二、启动流程

main.go 就干 4 件事，顺序不能乱：

```
1. config.LoadConfig()     — 读 .env 文件，没有就走环境变量，所有配置项有默认值
2. database.InitMySQL()    — 连 MySQL，GORM AutoMigrate 自动建/改表，不用手动跑 SQL
3. database.InitRedis()    — 连 Redis，Ping 一下确认通不通
4. SetupRouter().Run()     — 注册路由，监听端口（默认 8080）
```

改配置不用改代码，复制 `.env.example` 成 `.env` 改值就行。

---

## 三、请求怎么路由到 Handler 的

所有 API 以 `/api` 开头，按模块分 Group：

### 3.1 路由总表

| 模块 | 路径前缀 | JWT 鉴权 | 角色限制 | 对应 Handler 文件 |
|---|---|---|---|---|
| 认证 | `/api/auth` | 仅 `/me` 需要登录 | 无 | auth_handler.go |
| 医生 | `/api/doctors` | 写操作需要 | 管理员 | service_handler.go |
| 服务 | `/api/services` | 写操作需要 | 管理员 | service_handler.go |
| 宠物 | `/api/pets` | 全部需要 | 无（但 owner_id 隔离） | pet_handler.go |
| 排班 | `/api/schedules` | 全部需要 | 写操作限管理员/医生 | schedule_handler.go |
| 预约 | `/api/appointments` | 全部需要 | 确认/拒绝限医生/管理员 | appointment_handler.go |

### 3.2 具体路由清单

```
POST   /api/auth/register              — 注册（公开）
POST   /api/auth/login                 — 登录（公开）
GET    /api/auth/me                    — 当前用户信息 🔒

GET    /api/doctors                    — 医生列表（公开）
GET    /api/doctors/:id                — 医生详情（公开）
POST   /api/doctors                    — 新增医生 🔒 管理员
PUT    /api/doctors/:id                — 修改医生 🔒 管理员
DELETE /api/doctors/:id                — 删除医生 🔒 管理员

GET    /api/services                   — 服务列表（公开）
GET    /api/services/categories        — 服务分类（公开）
GET    /api/services/:id               — 服务详情（公开）
POST   /api/services                   — 新增服务 🔒 管理员
PUT    /api/services/:id               — 修改服务 🔒 管理员
DELETE /api/services/:id               — 删除服务 🔒 管理员

POST   /api/pets                       — 添加宠物 🔒
GET    /api/pets                       — 我的宠物列表 🔒
GET    /api/pets/:id                   — 宠物详情 🔒（校验归属）
PUT    /api/pets/:id                   — 修改宠物 🔒（校验归属）
DELETE /api/pets/:id                   — 删除宠物 🔒（校验归属）

GET    /api/schedules                  — 排班列表 🔒
GET    /api/schedules/:id              — 排班详情 🔒
GET    /api/schedules/doctor/:doctor_id — 医生某日排班 🔒（走 Redis 缓存）
POST   /api/schedules                  — 创建排班 🔒 管理员/医生
PUT    /api/schedules/:id              — 修改排班 🔒 管理员/医生
DELETE /api/schedules/:id              — 删除排班 🔒 管理员/医生

POST   /api/appointments               — 创建预约 🔒
GET    /api/appointments/my            — 我的预约 🔒
GET    /api/appointments/:id           — 预约详情 🔒
POST   /api/appointments/:id/cancel    — 取消预约 🔒
GET    /api/appointments               — 全部预约 🔒 医生/管理员
GET    /api/appointments/doctor/mine   — 医生的预约 🔒 医生/管理员
POST   /api/appointments/:id/confirm   — 确认预约 🔒 医生/管理员
POST   /api/appointments/:id/reject    — 拒绝预约 🔒 医生/管理员
```

🔒 = 需要 JWT，🔒 管理员 = JWT + RoleAuth(admin)

### 3.3 中间件执行链

Gin 的中间件是洋葱模型，每个请求按以下顺序穿过：

```
请求进来
  → gin.Default() 自带的 Logger + Recovery
  → CORS（允许跨域）
  → JWTAuth()（仅挂在需要的路由组上）
      → 检查 Authorization: Bearer xxx 头
      → 解析 JWT，把 userID / phone / role / username 写进 gin.Context
      → 不合法直接 401 Abort，不到 handler
  → RoleAuth(roles...)（仅挂在需要角色控制的路由组上）
      → 从 Context 取 role，判断是否在允许列表里
      → 不在列表直接 403 Abort
  → 你的 Handler 函数
```

**关键点**：JWTAuth 在 Context 里写的是 `userID`、`phone`、`role`、`username` 这四个 key，handler 里用 `c.Get("userID")` 取。

---

## 四、JWT 鉴权机制

### 4.1 签发

登录/注册成功后调用 `middleware.GenerateToken()`，把以下信息塞进 JWT Payload：

| 字段 | 含义 |
|---|---|
| user_id | 用户 ID |
| phone | 手机号 |
| role | 角色：user / doctor / admin |
| username | 昵称 |
| exp | 过期时间（默认 72 小时，环境变量 JWT_EXPIRE_HOURS 可调） |
| iss | 固定 "pet-clinic" |

签名算法 HS256，密钥从 `JWT_SECRET` 环境变量读。

### 4.2 解析

JWTAuth 中间件从 `Authorization: Bearer <token>` 头取 token，用 `jwt.ParseWithClaims` 解析。

- token 缺失 → 10002 未授权
- 格式不对 → 20004 Token无效
- 过期 → 20005 Token已过期
- 签名错 → 20004 Token无效

### 4.3 三种角色

| 角色 | 能干什么 |
|---|---|
| user | 管理自己的宠物、预约就诊、取消预约 |
| doctor | 同 user + 查看/管理排班、确认/拒绝预约 |
| admin | 同 doctor + 管理医生账号、管理服务项目 |

角色是在 User 表的 role 字段存的，创建 Doctor 时会自动把关联 User 的 role 改成 "doctor"。

---

## 五、Redis 缓存机制

### 5.1 缓存了什么

只缓存了一样东西：**某医生某日的排班列表**。

Key 格式：`schedule:doctor:{doctorID}:date:{yyyy-MM-dd}`

比如 `schedule:doctor:3:date:2026-06-26`

Value：该医生该日所有活跃排班的 JSON 数组，包含 booked 数量等全部字段。

TTL：24 小时。

### 5.2 读缓存流程

只有 `GET /api/schedules/doctor/:doctor_id` 这一个接口走缓存：

```
1. 先查 Redis（GetCachedTodaySchedule）
2. 命中 → 直接返回，不碰 MySQL
3. 未命中 → 查 MySQL，如果结果非空就写回 Redis，返回
```

### 5.3 缓存失效时机

以下 6 个操作会在事务提交后删除对应缓存 key：

| 操作 | 失效的 Key |
|---|---|
| 创建排班 | `schedule:doctor:{新排班的doctorID}:date:{日期}` |
| 修改排班 | 同上 |
| 删除排班 | 同上 |
| 预约成功 | `schedule:doctor:{预约的doctorID}:date:{预约日期}`（booked 变了） |
| 拒绝预约 | 同上 |
| 取消预约 | 同上 |

**失效在 tx.Commit() 之后执行**，不会出现"先删缓存再事务回滚"导致的不一致。

### 5.4 一致性模型

最终一致。并发下可能出现秒级不一致窗口（管理员改排班和用户同时查排班），下次请求自动修复。如果你要改缓存策略，关注 `database/redis.go` 里的 `CacheTodaySchedule` / `GetCachedTodaySchedule` / `InvalidateScheduleCache` 三个函数。

---

## 六、MySQL 表结构

GORM AutoMigrate 自动建表，不用手动跑 DDL。以下是大致的表关系：

```
User (用户表)
 ├── 1:1 ── Doctor (医生表，通过 user_id 关联)
 ├── 1:N ── Pet (宠物表，通过 owner_id 关联)
 └── 1:N ── Appointment (预约表，通过 user_id 关联)

Doctor (医生表)
 └── 1:N ── Schedule (排班表，通过 doctor_id 关联)
 └── 1:N ── Appointment (预约表，通过 doctor_id 关联)

Service (服务项目表，独立)
 └── 1:N ── Appointment (预约表，通过 service_id 关联)

Pet (宠物表)
 └── 1:N ── Appointment (预约表，通过 pet_id 关联)

Schedule (排班表)
 └── 1:N ── Appointment (预约表，通过 schedule_id 关联)
```

### 6.1 各表字段速查

**users**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| phone | varchar(20) UNIQUE | 登录手机号 |
| password | varchar(255) | BCrypt 加密，JSON 输出时隐藏（`json:"-"`） |
| username | varchar(50) | 昵称 |
| role | varchar(20) | user / doctor / admin |
| avatar | varchar(255) | 头像 URL |

**doctors**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| user_id | uint INDEX | 关联 users.id |
| name | varchar(50) | 医生姓名 |
| title | varchar(50) | 职称（主任/副主任等） |
| specialty | varchar(100) | 专科方向 |
| is_active | bool | 是否在职 |

**services**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| name | varchar(100) | 服务名（如"狂犬疫苗"） |
| category | varchar(50) | 分类（疫苗/绝育/体检等） |
| price | decimal(10,2) | 价格 |
| duration | int | 预计耗时（分钟） |
| is_active | bool | 是否上架 |

**pets**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| owner_id | uint INDEX | 归属用户 ID，**所有操作都校验此字段** |
| name | varchar(50) | 宠物名字 |
| species | varchar(30) | 种类（猫/狗/兔...） |
| age | int | 年龄 |
| gender | varchar(10) | 性别 |
| breed | varchar(50) | 品种 |

**schedules**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| doctor_id | uint INDEX | 哪个医生的排班 |
| date | varchar(10) INDEX | 日期，格式 yyyy-MM-dd |
| start_time | varchar(5) | 开始时间，格式 HH:mm |
| end_time | varchar(5) | 结束时间 |
| max_slots | int | 该时段最大预约数 |
| booked | int | 已预约数（**原子更新，防并发超卖**） |
| is_active | bool | 是否生效 |

**appointments**

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| user_id | uint INDEX | 预约用户 |
| doctor_id | uint INDEX | 预约医生 |
| schedule_id | uint INDEX | 关联排班 |
| service_id | uint INDEX | 服务项目 |
| pet_id | uint INDEX | 关联宠物（可为 0，兼容旧数据） |
| pet_name | varchar(50) | 宠物名字（冗余，从 Pet 或请求体填入） |
| pet_type | varchar(30) | 宠物种类（冗余） |
| date | varchar(10) INDEX | 预约日期 |
| start_time / end_time | varchar(5) | 预约时段（冗余，从 Schedule 填入） |
| status | varchar(20) INDEX | pending / confirmed / rejected / cancelled / completed |
| note | text | 用户备注 |
| reject_reason | text | 拒绝原因 |

### 6.2 冗余字段的设计意图

`pet_name`、`pet_type`、`date`、`start_time`、`end_time` 在 Appointment 里是冗余存的。原因：即使原排班被删改、宠物信息被修改，历史预约记录仍保留创建时的快照，不会出现"改了排班时间后看不懂旧预约"的问题。

---

## 七、预约时段冲突检测 — 最核心的逻辑

这段逻辑在 `appointment_handler.go` 的 `CreateAppointment` 函数里，是整个系统最复杂的一条链路。我按执行顺序一步步拆：

### 7.1 完整校验链

```
① 参数绑定失败 → 10001 参数错误

② 解析宠物信息（resolvePetInfo）
   ├─ 传了 pet_id → 查 Pet 表，不存在 → 50001
   │                   存在但 owner_id ≠ 当前用户 → 50002（防越权）
   │                   通过 → 用 pet.Name / pet.Species 填入
   └─ 没传 pet_id → pet_name 必填，否则 → 50003

③ 查排班 → 不存在 → 30002

④ 排班的 doctor_id ≠ 请求的 doctor_id → 40006

⑤ 排班未激活 → 30002

⑥ 排班已满（booked ≥ max_slots）→ 40005
   注意：这里是前置快速拦截，后面事务里还有二次检查

⑦ 查服务项目 → 不存在 → 40004

⑧ 时段冲突检测（checkTimeConflict）→ 冲突 → 40001
   SQL 判定逻辑：同一个医生同一天，存在状态为 pending 或 confirmed 的预约，
   且时间段有重叠

⑨ 开事务：
   ├─ 写入 Appointment 记录
   ├─ 原子更新 booked + 1（WHERE booked < max_slots）
   │   └─ RowsAffected == 0 → 40005 该时段预约已满（二次防超卖）
   └─ 提交事务

⑩ 失效 Redis 缓存
```

### 7.2 时段冲突的 SQL 逻辑

`checkTimeConflict` 函数的 WHERE 条件拆解：

```sql
WHERE doctor_id = ?
  AND date = ?
  AND status IN ('pending', 'confirmed')
  AND (
    (start_time <= 新结束时间 AND end_time > 新开始时间)   -- 已有预约的区间包含了新预约的起点
    OR
    (start_time < 新结束时间 AND end_time >= 新开始时间)   -- 已有预约的区间包含了新预约的终点
  )
```

直觉理解：**任何两个时间段有交集，就算冲突**。只检查 pending 和 confirmed 状态，rejected/cancelled 的不算。

### 7.3 防超卖的双重保险

预约满了怎么判断？两道锁：

1. **内存预检**（第⑥步）：`schedule.Booked >= schedule.MaxSlots`，快速拦截，减少数据库压力
2. **SQL 原子锁**（第⑨步事务内）：
   ```sql
   UPDATE schedules SET booked = booked + 1 WHERE id = ? AND booked < max_slots
   ```
   如果 `RowsAffected == 0`，说明被别人抢先约满了，事务回滚，返回 40005。

这样即使两个请求同时通过内存预检（都读到 booked=0, max_slots=1），SQL 层面的行锁也只有一个能成功，另一个 RowsAffected=0 被拦截。

### 7.4 预约状态流转

```
  用户创建预约
       │
       ▼
   [pending]
    ╱       ╲
   ▼         ▼
[confirmed] [rejected]  ← 医生操作
   │          │
   ▼          ▼
[cancelled] [cancel]    ← 用户取消（pending 或 confirmed 都可取消）
[completed]             ← 就诊完成（需手动改状态）
```

拒绝和取消时都会**原子减 booked**（`WHERE booked > 0` 防负数），并失效缓存。

---

## 八、统一响应格式

所有接口返回同一结构：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

错误时 `data` 不出现。错误码按模块分段：

| 范围 | 模块 | 示例 |
|---|---|---|
| 10001-10006 | 通用 | 10001 参数错误, 10003 无权限 |
| 20001-20005 | 认证 | 20003 密码错误, 20005 Token过期 |
| 30001-30003 | 排班 | 30003 该时段已有排班 |
| 40001-40006 | 预约 | 40001 预约时段冲突, 40005 预约已满 |
| 50001-50003 | 宠物 | 50002 只能预约自己名下的宠物 |

新增错误码去 `common/response.go` 加常量 + errorMessages 映射，**不要在 handler 里硬编码消息字符串**。

---

## 九、改代码时你应该知道的

### 加新接口
1. handler 写在 `handlers/` 下对应文件
2. 路由在 `router.go` 的对应 Group 里挂
3. 需要鉴权就加 `middleware.JWTAuth()`，需要角色控制就加 `middleware.RoleAuth()`

### 加新表
1. 在 `models/models.go` 加结构体
2. 在 `database/db.go` 的 `AutoMigrate` 调用里注册
3. 重启服务自动建表

### 加新错误码
1. 在 `common/response.go` 加 `ErrXxx ErrorCode = NNNNN` 常量
2. 在 `errorMessages` map 里加对应消息
3. handler 里用 `common.ErrorResponse(common.ErrXxx)`

### 通用校验模式
- 查存在性：`common.CheckExists(id, &dest, common.ErrXxxNotFound)`
- 查存在 + 归属：`common.CheckOwned(id, ownerID, &dest, getter, notFound, notOwned)`
- 写响应 + Abort：`common.Ensure(c, code, httpStatus)`

---

## 十、技术栈版本

| 组件 | 包 | 说明 |
|---|---|---|
| Web 框架 | github.com/gin-gonic/gin | HTTP 路由 + 中间件 |
| ORM | gorm.io/gorm + gorm.io/driver/mysql | MySQL 操作 + 自动迁移 |
| Redis | github.com/redis/go-redis/v9 | 排班缓存 |
| JWT | github.com/golang-jwt/jwt/v5 | HS256 签名 |
| 密码加密 | golang.org/x/crypto/bcrypt | BCrypt |
| 环境变量 | github.com/joho/godotenv | 读 .env |
| CORS | github.com/gin-contrib/cors | 跨域 |
