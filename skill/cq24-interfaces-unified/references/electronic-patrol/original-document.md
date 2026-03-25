# API接口调用说明V1.0

| 版本 | 编写人 | 日期 |
| --- | --- | --- |
| v1.0 | 陈庆 | 2020-06-16 |
| v1.1 | 陈庆 | 2021-01-18 |
| v1.2 | 陈庆 | 2021-04-19 |

## 目录

- 1 约定
  - 1.1 通用返回对象
  - 1.2 通用分页对象
- 2 接口
  - 2.1 登录接口
  - 2.2 部门接口
  - 2.3 权限组管理
  - 2.4 管理员账号
  - 2.5 巡检员接口
  - 2.6 巡检点接口
  - 2.7 线路管理
  - 2.8 工作计划
  - 2.9 设备管理
  - 2.10 巡检记录
  - 2.11 考核记录
  - 2.12 任务班次

---

# 约定

所有时间使用带毫秒的时间戳

除登录接口外，其它接口都需要在HTTP HEAD的Authorization添加Token值

## 通用返回对象

返回为以下3个字段，后面接口只对data字段进行说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| code | Int | 返回码 |
| msg | String | 错误信息 |
| data | Object | 数据 |

## 通用分页对象

查询带分页的接口 后面接口只对rows字段进行说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| total | int | 数据总数 |
| pageNum | int | 当前页码 |
| pageSize | int | 当前页大小 |
| totalPage | int | 总页数 |
| rows | Object | 数据对象 |

---

# 接口

## 登录接口

地址： GET /api/v1/login

参数

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| username | String | 是 | 登录名 |
| password | String | 是 | MD5加密后的密码 |

返回

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | String | Token值 |

## 部门接口

### 获取部门列表

地址：GET /api/v1/areas/query

参数： 无

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | 部门ID |
| name | String | 部门名称 |
| parentId | Long | 父部门ID |
| children | Object | 子部门列表 |

### 添加部门

地址: POST /api/v1/areas/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 部门名称 |
| parentId | Long | 是 | 上级部门ID |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 部门ID |

### 修改部门

地址: POST /api/v1/areas/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 部门ID |
| name | String | 是 | 部门名称 |
| parentId | Long | 是 | 上级部门ID |

### 删除部门

地址: POST /api/v1/areas/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 部门ID |

## 权限组管理

### 查询权限组

地址：GET /api/v1/roles/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 否 | 名称 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| name | String | 名称 |
| description | Long | 描述 |
| createTime | Long | 创建时间 |

### 添加权限组

地址: POST /api/v1/roles/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 名称 |
| description | String | 否 | 描述 |
| privilegeIds | Long[] | 是 | 权限ID列表 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 权限组ID |

### 修改权限组

地址: POST /api/v1/roles/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | ID |
| name | String | 是 | 名称 |
| description | String | 否 | 描述 |
| privilegeIds | Long[] | 是 | 权限ID列表 |

### 删除权限组

地址: POST /api/v1/roles/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 权限组ID |

### 获取所有权限列表

地址：GET /api/v1/roles/privileges

参数： 无

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | Long | 权限码 |
| name | String | 名称 |
| url | String | (无用) |
| code | String | 权限码等同于ID |
| parentId | Long | 上级权限码 |

### 获取所有权限组下的权限列表

地址：GET /api/v1/roles/{id}/privileges

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 权限组ID |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | Long | 权限码 |
| name | String | 名称 |
| url | String | (无用) |
| code | String | 权限码等同于ID |
| parentId | Long | 上级权限码 |

## 管理员账号

### 查询管理员

地址：GET /api/v1/users/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| nickname | String | 否 | 姓名 |
| areaIds | Long[] | 否 | 部门列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页)：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| username | String | 账号 |
| nickname | String | 姓名 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| createTime | Long | 创建时间 |
| loginTime | Long | 登录时间 |
| roleIds | Long[] | 权限组ID列表 |
| roleNames | String[] | 权限组名称列表 |

### 添加管理员

地址: POST /api/v1/users/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| username | String | 是 | 账号 |
| nickname | String | 是 | 姓名 |
| areaId | Long | 是 | 部门ID |
| roleIds | Long[] | 是 | 权限组ID列表 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 管理员ID |

### 修改管理员

地址: POST /api/v1/users/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 管理员ID |
| username | String | 是 | 账号 |
| nickname | String | 是 | 姓名 |
| areaId | Long | 是 | 部门ID |
| roleIds | Long[] | 是 | 权限组ID列表 |

### 删除管理员

地址: POST /api/v1/users/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 管理员ID |

## 巡检员接口

### 查询巡检员

地址：GET /api/v1/patrolmans/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 否 | 名称 |
| card | String | 否 | 人员卡号 |
| areaIds | Long[] | 否 | 区域列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页)：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | 人员ID |
| name | String | 人员名称 |
| card | String | 人员卡号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| remark | String | 备注 |

### 添加巡检员

地址: POST /api/v1/patrolmans/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 人员名称 |
| card | String | 是 | 人员卡号 |
| areaId | Long | 是 | 部门ID |
| remark | String | 否 | 备注 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 人员ID |

### 修改巡检员

地址: POST /api/v1/patrolmans/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 人员ID |
| name | String | 是 | 人员名称 |
| card | String | 是 | 人员卡号 |
| areaId | Long | 是 | 部门ID |
| remark | String | 否 | 备注 |

### 删除巡检员

地址: POST /api/v1/patrolmans/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 巡检员ID |

## 巡检点接口

### 查询巡检点

地址：GET /api/v1/checkpoints/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 否 | 名称 |
| card | String | 否 | 卡号 |
| areaIds | Long[] | 否 | 部门ID列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页)：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| name | String | 名称 |
| card | String | 卡号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| remark | String | 备注 |

### 添加巡检点

地址: POST /api/v1/checkpoints/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 名称 |
| card | String | 是 | 卡号 |
| areaId | Long | 是 | 部门ID |
| remark | String | 否 | 备注 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 巡检点ID |

### 修改巡检点

地址: POST /api/v1/checkpoint/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | ID |
| name | String | 是 | 名称 |
| card | String | 是 | 卡号 |
| areaId | Long | 是 | 部门ID |
| remark | String | 否 | 备注 |

### 删除巡检点

地址: POST /api/v1/checkpoint/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 巡检员ID |

## 线路管理

### 查询线路

地址：GET /api/v1/lines/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 否 | 名称 |
| areaIds | Long[] | 否 | 部门列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页)：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| name | String | 名称 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |

### 添加线路

地址: POST /api/v1/lines/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 名称 |
| areaId | Long | 是 | 部门ID |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 线路ID |

### 修改线路

地址: POST /api/v1/lines/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | ID |
| name | String | 是 | 名称 |

### 删除线路

地址: POST /api/v1/lines/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 线路ID |

### 获取线路下的点位

地址：GET /api/v1/lines/{lineId}/checkpoints

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| lineId | Long | 是 | 线路ID @PathVariable |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| name | String | 名称 |
| card | String | 卡号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| remark | String | 备注 |

### 向线路中添加点位

地址：GET /api/v1/nodes/add

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| lineId | Long | 是 | 线路ID |
| checkpointIds | Long[] | 是 | 点位ID列表 |

### 移除线路中的点位

地址：GET /api/v1/nodes/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| lineId | Long | 是 | 线路ID |
| checkpointId | Long | 是 | 点位ID |

## 工作计划

### 查询工作计划

地址：GET /api/v1/plans/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 否 | 名称 |
| lineIds | Long[] | 否 | 线路ID列表 |
| startDate | Long | 否 | 开始日期(对比任务结束日期) |
| endDate | Long | 否 | 结束日期(对比任务开始日期) |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页)：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| name | String | 名称 |
| startDate | Long | 开始日期 |
| endDate | Long | 结束日期 |
| startTime | Long | 开始时间 |
| endTime | Long | 结束时间 |
| patrol | int | 巡逻一圈所需时间 |
| rest | int | 巡逻一圈后休息时间 |
| mon | boolean | 周一是否巡检 |
| tue | boolean | 周二是否巡检 |
| wed | boolean | 周三是否巡检 |
| thu | boolean | 周四是否巡检 |
| fri | boolean | 周五是否巡检 |
| sat | boolean | 周六是否巡检 |
| sun | boolean | 周天是否巡检 |
| lineId | Long | 线路Id |
| lineName | String | 线路名称 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| createTime | Long | 创建时间 |

### 添加计划

地址: POST /api/v1/plans/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 名称 |
| startDate | Long | 是 | 开始日期 |
| endDate | Long | 是 | 结束日期 |
| startTime | Long | 是 | 开始时间 |
| endTime | Long | 是 | 结束时间 |
| patrol | int | 是 | 巡逻一圈所需时间 |
| rest | int | 是 | 巡逻一圈休息时间 |
| lineId | Long | 是 | 线路ID |
| mon | boolean | 是 | 周一是否巡检 |
| tue | boolean | 是 | 周二是否巡检 |
| wed | boolean | 是 | 周三是否巡检 |
| thu | boolean | 是 | 周四是否巡检 |
| fri | boolean | 是 | 周五是否巡检 |
| sat | boolean | 是 | 周六是否巡检 |
| sun | boolean | 是 | 周天是否巡检 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 计划ID |

### 修改计划

地址: POST /api/v1/plans/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 计划ID |
| name | String | 是 | 名称 |
| startDate | Long | 是 | 开始日期 |
| endDate | Long | 是 | 结束日期 |
| startTime | Long | 是 | 开始时间 |
| endTime | Long | 是 | 结束时间 |
| patrol | int | 是 | 巡逻一圈所需时间 |
| rest | int | 是 | 巡逻一圈休息时间 |
| lineId | Long | 是 | 线路ID |
| mon | boolean | 是 | 周一是否巡检 |
| tue | boolean | 是 | 周二是否巡检 |
| wed | boolean | 是 | 周三是否巡检 |
| thu | boolean | 是 | 周四是否巡检 |
| fri | boolean | 是 | 周五是否巡检 |
| sat | boolean | 是 | 周六是否巡检 |
| sun | boolean | 是 | 周天是否巡检 |

### 删除计划

地址：GET /api/v1/plans/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 计划ID |

## 设备管理

### 查询设备

地址：GET /api/v1/devices/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 否 | 名称 |
| code | String | 否 | 设备号 |
| areaIds | Long | 否 | 部门ID列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页)：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| name | String | 名称 |
| code | String | 设备号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| patrolmanId | Long | 巡检员ID |
| patrolmanName | String | 巡检员名称 |
| patrolmanCard | String | 巡检员卡号 |
| remark | String | 备注 |

### 添加设备

地址: POST /api/v1/devices/add

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| name | String | 是 | 名称 |
| code | String | 是 | 设备号 |
| areaId | Long | 是 | 部门ID |
| patrolmanId | Long | 否 | 巡检员ID |
| remark | String | 否 | 备注 |

返回：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| data | Long | 设备ID |

### 修改设备

地址: POST /api/v1/devices/update

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 设备ID |
| name | String | 是 | 名称 |
| code | String | 是 | 设备号 |
| areaId | Long | 是 | 部门ID |
| patrolmanId | Long | 否 | 巡检员ID |
| remark | String | 否 | 备注 |

### 删除设备

地址：POST /api/v1/devices/remove

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| id | Long | 是 | 设备ID |

## 巡检记录

### 查询巡检记录

地址：GET /api/v1/checkpointLogs/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| startTime | Long | 是 | 开始时间 |
| endTime | Long | 是 | 结束时间 |
| patrolmanId | Long | 否 | 巡检员ID |
| checkpointId | Long | 否 | 巡检点ID |
| deviceId | Long | 否 | 设备ID |
| areaIds | Long | 否 | 部门ID列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页):

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| patrolmanId | Long | 巡检员ID |
| patrolmanName | String | 巡检员名称 |
| patrolmanCard | String | 巡检员卡号 |
| checkpointId | Long | 巡检点ID |
| checkpointName | String | 巡检点名称 |
| checkpointCard | String | 巡检点卡号 |
| deviceId | Long | 设备ID |
| deviceName | String | 设备名称 |
| deviceCode | String | 设备号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| createTime | Long | 读卡时间 |
| uploadTime | Long | 上传时间 |

## 考核记录

### 查询考核记录

地址：GET /api/v1/planChecks/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| startDate | Long | 是 | 开始时间(对比任务开始时间) |
| endDate | Long | 是 | 结束时间(对比任务开始时间) |
| checkpointId | Long | 否 | 巡检点ID |
| lineIds | Long[] | 否 | 线路ID列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回(分页):

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| startTime | Long | 任务开始时间 |
| endTime | Long | 任务结束时间 |
| arriveTime | Long | 签到时间 |
| arriveType | int | 状态 -1未到 0准时 |
| lineId | Long | 线路ID |
| lineName | String | 线路名称 |
| patrolmanId | Long | 巡检员ID |
| patrolmanName | String | 巡检员名称 |
| patrolmanCard | String | 巡检员卡号 |
| checkpointId | Long | 巡检点ID |
| checkpointName | String | 巡检点名称 |
| checkpointCard | String | 巡检点卡号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| remark | String | 备注 |

### 添加备注

地址: POST /api/v1/planChecks/setRemark

参数：HTTP 头：Content-Type application/json

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| ids | Long[] | 是 | 考核计划ID |
| remark | String | 否 | 备注 |

### 统计考核记录

地址：GET /api/v1/planChecks/report

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| startDate | Long | 是 | 开始时间(对比任务开始时间) |
| endDate | Long | 是 | 结束时间(对比任务开始时间) |
| checkpointId | Long | 否 | 巡检点ID |
| lineIds | Long[] | 否 | 线路ID列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| count | int | 任务总数 |
| arriveCount | Long | 准时总数 |
| notArriveCount | int | 未到总数 |
| lineId | Long | 线路ID |
| lineName | String | 线路名称 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |

## 任务班次

### 查询任务班次

地址：GET /api/v1/planJobs/query

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| startTime | Long | 是 | 开始时间(对比任务开始时间) |
| endTime | Long | 是 | 结束时间(对比任务开始时间) |
| lineIds | Long[] | 否 | 线路ID列表 |
| pageNum | int | 否 | 页码 默认1 |
| pageSize | int | 否 | 页大小 默认15 |

返回:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | Long | 任务班次ID |
| startTime | Long | 任务开始时间 |
| endTime | Long | 任务结束时间 |
| lineId | Long | 线路ID |
| lineName | String | 线路名称 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |
| planId | Long | 计划ID |
| planName | String | 计划名称 |
| count | int | 任务总数 |
| arriveCount | Long | 准时总数 |
| notArriveCount | int | 未到总数 |
| progress | double | 完成率 |

### 查询任务班次详情

地址：GET /api/v1/planJobs/planChecks

参数：

| 字段 | 类型 | 必须 | 说明 |
| --- | --- | --- | --- |
| planJobId | Long | 是 | 任务班次ID |

返回:

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | String | ID |
| startTime | Long | 任务开始时间 |
| endTime | Long | 任务结束时间 |
| arriveTime | Long | 签到时间 |
| arriveType | int | 状态 -1未到 0准时 |
| lineId | Long | 线路ID |
| lineName | String | 线路名称 |
| patrolmanId | Long | 巡检员ID |
| patrolmanName | String | 巡检员名称 |
| patrolmanCard | String | 巡检员卡号 |
| checkpointId | Long | 巡检点ID |
| checkpointName | String | 巡检点名称 |
| checkpointCard | String | 巡检点卡号 |
| areaId | Long | 部门ID |
| areaName | String | 部门名称 |