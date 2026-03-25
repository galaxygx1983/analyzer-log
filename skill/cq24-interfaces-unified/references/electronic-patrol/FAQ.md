# 电子巡更 API 接口文档

来源: API接口文档-电子巡更.docx

## 概述

本技能提供从文档中提取的问答对，用于快速查询接口信息。

**文档类型:** API 文档
**版本:** V1.0 (更新至 v1.2)
**受众:** 开发人员
**主题:** 认证授权、部门区域管理、权限组管理、管理员账号管理、巡检员管理、巡检点管理、线路管理、工作计划管理、设备管理、巡检记录管理、考核记录管理、任务班次管理

## 问答统计

**问答总数:** 52 对

**分类:**
- 配置规范: 2 个问题
- 通用约定: 2 个问题
- 登录接口: 2 个问题
- 部门管理: 4 个问题
- 权限组管理: 6 个问题
- 管理员管理: 4 个问题
- 巡检员管理: 4 个问题
- 巡检点管理: 4 个问题
- 线路管理: 7 个问题
- 工作计划管理: 4 个问题
- 设备管理: 4 个问题
- 巡检记录: 1 个问题
- 考核记录: 3 个问题
- 任务班次: 2 个问题

---

## 配置规范

### Q1. API 通用返回对象的结构是什么？

API 返回三个字段：code（返回码，Int 类型）、msg（错误信息，String 类型）、data（数据对象）。后续接口文档中仅描述 data 字段内容。

*来源: 1.1 通用返回对象*

### Q2. 分页响应对象的结构是什么？

分页响应包含：total（总数据量）、pageNum（当前页码）、pageSize（当前页大小）、totalPage（总页数）、rows（数据对象列表）。

*来源: 1.2 通用分页对象*

---

## 通用约定

### Q1. 本 API 使用什么认证方式？

除登录接口外，所有其他接口需在 HTTP 请求头的 Authorization 字段中添加 Token 值。

*来源: 1 约定*

### Q2. 本 API 使用什么时间格式？

所有时间戳均使用带毫秒的时间戳格式。

*来源: 1 约定*

---

## 登录接口

### Q1. 如何登录系统？

使用 GET 请求访问 `/api/v1/login`，参数：username（必填，登录名）、password（必填，MD5加密后的密码）。

*来源: 2.1 登录接口*

### Q2. 登录接口返回什么？

登录接口在 data 字段返回 Token 值。

*来源: 2.1 登录接口*

---

## 部门管理

### Q1. 如何获取部门列表？

使用 GET 请求访问 `/api/v1/areas/query`，无需参数。返回 id、name、parentId、children（子部门列表）。

*来源: 2.2.1 获取部门列表*

### Q2. 如何添加部门？

使用 POST 请求访问 `/api/v1/areas/add`，Content-Type 为 application/json。必填参数：name（部门名称）、parentId（上级部门ID）。返回新建部门 ID。

*来源: 2.2.2 添加部门*

### Q3. 如何修改部门信息？

使用 POST 请求访问 `/api/v1/areas/update`，Content-Type 为 application/json。必填参数：id、name、parentId。

*来源: 2.2.3 修改部门*

### Q4. 如何删除部门？

使用 POST 请求访问 `/api/v1/areas/remove`，参数：id（部门 ID，必填）。

*来源: 2.2.4 删除部门*

---

## 权限组管理

### Q1. 如何查询权限组？

使用 GET 请求访问 `/api/v1/roles/query`，可选参数：name（名称）。返回 id、name、description、createTime。

*来源: 2.3.1 查询权限组*

### Q2. 如何添加权限组？

使用 POST 请求访问 `/api/v1/roles/add`，Content-Type 为 application/json。必填参数：name、privilegeIds（权限 ID 数组）。可选参数：description。返回权限组 ID。

*来源: 2.3.2 添加权限组*

### Q3. 如何修改权限组？

使用 POST 请求访问 `/api/v1/roles/update`，Content-Type 为 application/json。必填参数：id、name、privilegeIds。可选参数：description。

*来源: 2.3.3 修改权限组*

### Q4. 如何删除权限组？

使用 POST 请求访问 `/api/v1/roles/remove`，参数：id（权限组 ID，必填）。

*来源: 2.3.4 删除权限组*

### Q5. 如何获取所有权限列表？

使用 GET 请求访问 `/api/v1/roles/privileges`，无需参数。返回 id、name、url、code（权限码，与 ID 相同）、parentId。

*来源: 2.3.5 获取所有权限列表*

### Q6. 如何获取指定权限组下的权限列表？

使用 GET 请求访问 `/api/v1/roles/{id}/privileges`，其中 id 为权限组 ID。返回权限详情，包括 id、name、url、code、parentId。

*来源: 2.3.6 获取所有权限组下的权限列表*

---

## 管理员管理

### Q1. 如何查询管理员？

使用 GET 请求访问 `/api/v1/users/query`，可选参数：nickname（姓名）、areaIds（部门列表）、pageNum（默认 1）、pageSize（默认 15）。返回分页结果，包含 id、username、nickname、areaId、areaName、createTime、loginTime、roleIds、roleNames。

*来源: 2.4.1 查询管理员*

### Q2. 如何添加管理员？

使用 POST 请求访问 `/api/v1/users/add`，Content-Type 为 application/json。必填参数：username、nickname、areaId、roleIds（权限组 ID 列表）。返回管理员 ID。

*来源: 2.4.2 添加管理员*

### Q3. 如何修改管理员信息？

使用 POST 请求访问 `/api/v1/users/update`，Content-Type 为 application/json。必填参数：id、username、nickname、areaId、roleIds。

*来源: 2.4.3 修改管理员*

### Q4. 如何删除管理员？

使用 POST 请求访问 `/api/v1/users/remove`，参数：id（管理员 ID，必填）。

*来源: 2.4.4 删除管理员*

---

## 巡检员管理

### Q1. 如何查询巡检员？

使用 GET 请求访问 `/api/v1/patrolmans/query`，可选参数：name、card（人员卡号）、areaIds（区域列表）、pageNum（默认 1）、pageSize（默认 15）。返回分页结果，包含 id、name、card、areaId、areaName、remark。

*来源: 2.5.1 查询巡检员*

### Q2. 如何添加巡检员？

使用 POST 请求访问 `/api/v1/patrolmans/add`，Content-Type 为 application/json。必填参数：name、card、areaId。可选参数：remark。返回人员 ID。

*来源: 2.5.2 添加巡检员*

### Q3. 如何修改巡检员信息？

使用 POST 请求访问 `/api/v1/patrolmans/update`，Content-Type 为 application/json。必填参数：id、name、card、areaId。可选参数：remark。

*来源: 2.5.3 修改巡检员*

### Q4. 如何删除巡检员？

使用 POST 请求访问 `/api/v1/patrolmans/remove`，参数：id（巡检员 ID，必填）。

*来源: 2.5.4 删除巡检员*

---

## 巡检点管理

### Q1. 如何查询巡检点？

使用 GET 请求访问 `/api/v1/checkpoints/query`，可选参数：name、card、areaIds、pageNum（默认 1）、pageSize（默认 15）。返回分页结果，包含 id、name、card、areaId、areaName、remark。

*来源: 2.6.1 查询巡检点*

### Q2. 如何添加巡检点？

使用 POST 请求访问 `/api/v1/checkpoints/add`，Content-Type 为 application/json。必填参数：name、card、areaId。可选参数：remark。返回巡检点 ID。

*来源: 2.6.2 添加巡检点*

### Q3. 如何修改巡检点？

使用 POST 请求访问 `/api/v1/checkpoint/update`，Content-Type 为 application/json。必填参数：id、name、card、areaId。可选参数：remark。

*来源: 2.6.3 修改巡检点*

### Q4. 如何删除巡检点？

使用 POST 请求访问 `/api/v1/checkpoint/remove`，参数：id（巡检点 ID，必填）。

*来源: 2.6.4 删除巡检点*

---

## 线路管理

### Q1. 如何查询线路？

使用 GET 请求访问 `/api/v1/lines/query`，可选参数：name、areaIds、pageNum（默认 1）、pageSize（默认 15）。返回分页结果，包含 id、name、areaId、areaName。

*来源: 2.7.1 查询线路*

### Q2. 如何添加线路？

使用 POST 请求访问 `/api/v1/lines/add`，Content-Type 为 application/json。必填参数：name、areaId。返回线路 ID。

*来源: 2.7.2 添加线路*

### Q3. 如何修改线路？

使用 POST 请求访问 `/api/v1/lines/update`，Content-Type 为 application/json。必填参数：id、name。

*来源: 2.7.3 修改线路*

### Q4. 如何删除线路？

使用 POST 请求访问 `/api/v1/lines/remove`，参数：id（线路 ID，必填）。

*来源: 2.7.4 删除线路*

### Q5. 如何获取线路下的点位？

使用 GET 请求访问 `/api/v1/lines/{lineId}/checkpoints`，其中 lineId 为线路 ID。返回点位列表，包含 id、name、card、areaId、areaName、remark。

*来源: 2.7.5 获取线路下的点位*

### Q6. 如何向线路中添加点位？

使用 GET 请求访问 `/api/v1/nodes/add`，参数：lineId（线路 ID，必填）、checkpointIds（点位 ID 列表，必填）。

*来源: 2.7.6 向线路中添加点位*

### Q7. 如何移除线路中的点位？

使用 GET 请求访问 `/api/v1/nodes/remove`，参数：lineId（线路 ID，必填）、checkpointId（点位 ID，必填）。

*来源: 2.7.7 移除线路中的点位*

---

## 工作计划管理

### Q1. 如何查询工作计划？

使用 GET 请求访问 `/api/v1/plans/query`，可选参数：name、lineIds、startDate、endDate、pageNum、pageSize。返回分页结果，包含 id、name、startDate、endDate、startTime、endTime、patrol、rest、mon-sun、lineId、lineName、areaId、areaName、createTime。

*来源: 2.8.1 查询工作计划*

### Q2. 如何添加工作计划？

使用 POST 请求访问 `/api/v1/plans/add`，Content-Type 为 application/json。必填参数：name、startDate、endDate、startTime、endTime、patrol、rest、lineId、mon-sun。返回计划 ID。

*来源: 2.8.2 添加计划*

### Q3. 如何修改工作计划？

使用 POST 请求访问 `/api/v1/plans/update`，Content-Type 为 application/json。必填参数：id、name、startDate、endDate、startTime、endTime、patrol、rest、lineId、mon-sun。

*来源: 2.8.3 修改计划*

### Q4. 如何删除工作计划？

使用 GET 请求访问 `/api/v1/plans/remove`，参数：id（计划 ID，必填）。

*来源: 2.8.4 删除计划*

---

## 设备管理

### Q1. 如何查询设备？

使用 GET 请求访问 `/api/v1/devices/query`，可选参数：name、code（设备号）、areaIds、pageNum、pageSize。返回分页结果，包含 id、name、code、areaId、areaName、patrolmanId、patrolmanName、patrolmanCard、remark。

*来源: 2.9.1 查询设备*

### Q2. 如何添加设备？

使用 POST 请求访问 `/api/v1/devices/add`，Content-Type 为 application/json。必填参数：name、code、areaId。可选参数：patrolmanId、remark。返回设备 ID。

*来源: 2.9.2 添加设备*

### Q3. 如何修改设备？

使用 POST 请求访问 `/api/v1/devices/update`，Content-Type 为 application/json。必填参数：id、name、code、areaId。可选参数：patrolmanId、remark。

*来源: 2.9.3 修改设备*

### Q4. 如何删除设备？

使用 POST 请求访问 `/api/v1/devices/remove`，参数：id（设备 ID，必填）。

*来源: 2.9.4 删除设备*

---

## 巡检记录

### Q1. 如何查询巡检记录？

使用 GET 请求访问 `/api/v1/checkpointLogs/query`，必填参数：startTime、endTime。可选参数：patrolmanId、checkpointId、deviceId、areaIds、pageNum、pageSize。返回分页结果，包含 id、patrolmanId、patrolmanName、patrolmanCard、checkpointId、checkpointName、checkpointCard、deviceId、deviceName、deviceCode、areaId、areaName、createTime、uploadTime。

*来源: 2.10.1 查询巡检记录*

---

## 考核记录

### Q1. 如何查询考核记录？

使用 GET 请求访问 `/api/v1/planChecks/query`，必填参数：startDate、endDate。可选参数：checkpointId、lineIds、pageNum、pageSize。返回分页结果，包含 id、startTime、endTime、arriveTime、arriveType（-1未到，0准时）、lineId、lineName、patrolmanId、patrolmanName、patrolmanCard、checkpointId、checkpointName、checkpointCard、areaId、areaName、remark。

*来源: 2.11.1 查询考核记录*

### Q2. 如何添加考核备注？

使用 POST 请求访问 `/api/v1/planChecks/setRemark`，Content-Type 为 application/json。必填参数：ids（考核计划 ID 数组）。可选参数：remark。

*来源: 2.11.2 添加备注*

### Q3. 如何统计考核记录？

使用 GET 请求访问 `/api/v1/planChecks/report`，必填参数：startDate、endDate。可选参数：checkpointId、lineIds、pageNum、pageSize。返回 count（任务总数）、arriveCount（准时总数）、notArriveCount（未到总数）、lineId、lineName、areaId、areaName。

*来源: 2.11.3 统计考核记录*

---

## 任务班次

### Q1. 如何查询任务班次？

使用 GET 请求访问 `/api/v1/planJobs/query`，必填参数：startTime、endTime。可选参数：lineIds、pageNum、pageSize。返回 id、startTime、endTime、lineId、lineName、areaId、areaName、planId、planName、count、arriveCount、notArriveCount、progress。

*来源: 2.12.1 查询任务班次*

### Q2. 如何查询任务班次详情？

使用 GET 请求访问 `/api/v1/planJobs/planChecks`，参数：planJobId（任务班次 ID，必填）。返回 id、startTime、endTime、arriveTime、arriveType、lineId、lineName、patrolmanId、patrolmanName、patrolmanCard、checkpointId、checkpointName、checkpointCard、areaId、areaName。

*来源: 2.12.2 查询任务班次详情*

---

## 参考文档

完整接口文档请参阅：[original-document.md](references/original-document.md)