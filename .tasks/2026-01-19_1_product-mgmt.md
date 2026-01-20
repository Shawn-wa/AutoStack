# 背景
文件名：2026-01-19_1
创建于：2026-01-19_15:30:00
创建者：Assistant
主分支：master
任务分支：task/product-mgmt_2026-01-19_1
Yolo模式：Ask

# 任务描述
1. 新增本地产品管理模块：处理系统本地产品管理（SKU基础信息，平台listing配对本地SKU管理）
2. 平台listing信息：根据Ozon产品接口拉取产品数据（包括主要listing信息和SKU子信息，库存，售价等）
3. 订单汇总功能：出单各三方SKU（包含配对本地SKU关系）的总量，需要根据订单状态来分组统计

# 项目概览
- Backend: Go (Gin, Gorm)
- Frontend: Vue 3 (Vite, TS)
- Database: MySQL

⚠️ 警告：永远不要修改此部分 ⚠️
[RIPER-5 协议摘要]
- 模式声明：必须在响应开头声明模式 [MODE: NAME]。
- 默认模式：RESEARCH。
- 创新模式：评估多种方案。
- 规划模式：详细规范，清单。
- 执行模式：严格遵循计划，更新任务文件。
- 审查模式：验证实施，偏差报告。
⚠️ 警告：永远不要修改此部分 ⚠️

# 分析
- **产品模块缺失**：当前没有本地产品管理模块，需要新建 `backend/internal/modules/product`。
- **Ozon 接口不足**：现有 Ozon Client 仅支持订单相关操作，需增加产品列表和详情接口。
- **关联关系**：需要建立 Local SKU 与 Platform SKU 的映射关系表。
- **订单统计**：需要新的 API 来聚合订单数据，并结合映射关系展示 Local SKU 的销量。

# 提议的解决方案
1. **数据模型**：
   - `Product` (本地产品): 基础信息 (SKU, Name, Image, Cost)。
   - `PlatformProduct` (平台产品): 从平台同步的原始数据 (PlatformSKU, Name, Stock, Price, Status)。
   - `ProductMapping` (关联关系): PlatformSKU + PlatformAuthID -> LocalSKU。

2. **Ozon 集成**：
   - 增加 `/v2/product/list` 和 `/v2/product/info/list` 接口调用。
   - 实现同步逻辑：拉取 -> Upsert `PlatformProduct`。

3. **订单统计**：
   - 新增 `GET /api/v1/order/stats/summary`。
   - 逻辑：聚合 `order_items`，Left Join `product_mappings`，Group by Status & SKU。

4. **前端实现**：
   - 新增 `Product` 模块。
   - 页面：本地产品列表（CRUD）、平台产品列表（查看+关联）、统计报表页面。

# 实施清单
1. Backend: Create `backend/internal/modules/product/model.go` with `Product`, `PlatformProduct`, `ProductMapping`.
2. Backend: Implement Ozon API methods in `backend/internal/modules/apiClient/platform/ozon/product.go` and `types.go`.
3. Backend: Create `backend/internal/modules/product/service.go` (CRUD, Sync, Mapping).
4. Backend: Create `backend/internal/modules/product/handler.go` and `dto.go`.
5. Backend: Register Product routes in `backend/internal/app/server.go`.
6. Backend: Add `GetOrderSummary` to `backend/internal/modules/order/service.go` and handler/route.
7. Frontend: Create `frontend/src/modules/product` directory structure.
8. Frontend: Implement `api/index.ts` for product module.
9. Frontend: Create `LocalProducts.vue` page.
10. Frontend: Create `PlatformProducts.vue` page with sync and mapping features.
11. Frontend: Create `OrderSummary.vue` page.
12. Frontend: Register routes in `frontend/src/router/index.ts`.

# 当前执行步骤："订单汇总按本地SKU合并显示"

# 任务进度
[2026-01-19 16:30:00]
- 已修改：
  - backend/internal/modules/order/dto.go：新增 OrderSummaryStatusDetail，修改 OrderSummaryItem 结构
  - backend/internal/modules/order/service.go：重写 GetOrderSummary 按本地SKU聚合逻辑
  - frontend/src/modules/product/api/index.ts：更新 OrderSummaryItem 类型定义
  - frontend/src/modules/product/pages/OrderSummary.vue：实现展开行UI显示各状态明细
- 更改：订单汇总页面按本地SKU合并显示，支持展开查看各状态销量和销售额
- 原因：用户需求按本地SKU合并显示行，可展开查看各状态的订单汇总信息
- 阻碍因素：无
- 状态：成功

[2026-01-19 16:45:00]
- 已修改：backend/internal/modules/product/service.go
- 更改：修复 SyncPlatformProducts 方法，在创建客户端前先解密凭证
- 原因：auth.Credentials 是加密存储的，需要先调用 order.Decrypt() 解密后再传给 createOzonClient
- 阻碍因素：无
- 状态：成功

[2026-01-19 17:00:00]
- 已修改：
  - backend/internal/modules/product/model.go：新增 PlatformSyncTask 模型和任务类型/状态枚举
  - backend/internal/modules/product/service.go：添加 CreateSyncTask、ExecuteSyncTask、ProcessPendingTasks、CleanOldTasks 方法
  - backend/internal/modules/product/handler.go：SyncPlatformProducts 改为创建任务记录
  - backend/internal/scheduler/scheduler.go：添加每5分钟扫描任务和每天1:20清理任务
  - backend/internal/app/server.go：AutoMigrate 添加 PlatformSyncTask
- 更改：实现平台同步任务队列机制，支持状态幂等隔离、最多5次重试、定时清理3个月前记录
- 原因：用户需求追踪同步任务状态，支持重试和清理
- 阻碍因素：无
- 状态：未确认

# 最终审查
