# vela-radar
Active scanning of network assets  
利用ssoc agent实现网络资产主动扫描  


## 更新日志
2023-11-10 &emsp; v0.1.0 &emsp; 初始测试版本  
2023-11-13 &emsp; v0.1.1 &emsp; 资产扫描功能优化, 加入web指纹探测功能  
2023-11-15 &emsp; v0.1.2 &emsp; 优化扫描数据的处理, 对接后端上报接口的数据结构  
2023-11-17 &emsp; v0.2.0 &emsp; 实现远程调用的内部API, 实时进度显示  
2023-11-20 &emsp; v0.2.1 &emsp; 实现禁止扫描的白名单ip设置  
2023-11-29 &emsp; v0.3.0 &emsp; 实现web站点截图功能  
2023-11-30 &emsp; v0.3.1 &emsp; 实现通过内部tunnel上报, web站点截图功能优化  
2023-12-20 &emsp; v0.3.2 &emsp; web指纹加载逻辑升级, 支持自定义三方指纹库动态加载和更新  
2023-12-25 &emsp; v0.4.0 &emsp; 支持设置扫描时间段(定时暂停与开始, beta)  
2023-12-25 &emsp; v0.4.1 &emsp; 优化任务计数逻辑, 提升扫描进度百分比的精确度  
2023-12-25 &emsp; v0.4.2 &emsp; 修复扫描只一个大段网段的某一个端口时, 导致的任务异常退出问题  
2023-12-25 &emsp; v0.4.3 &emsp; 修复FingerPrint子任务调度相关问题  
2024-01-02 &emsp; v0.4.4 &emsp; 内部HTTP API支持扫描排除时间段设置  
2024-01-03 &emsp; v0.4.5 &emsp; 修复扫描速率参数设置问题  
2024-01-05 &emsp; v0.4.5 &emsp; 优化扫描参数传入    
2024-01-06 &emsp; v0.4.6 &emsp; 优化扫描排除时间功能,现支持更加智能地排除开盘时间       
2024-01-06 &emsp; v0.4.6 &emsp; 整理完善文档       
2024-01-08 &emsp; v0.4.7 &emsp; 放宽传输层最大发包速率上限  
2024-01-08 &emsp; v0.4.7 &emsp; 修复内置TOP5000端口号部分错误   
2024-01-09 &emsp; v0.4.8 &emsp; 修复对监视器的通道关闭的处理相关错误  
2024-01-10 &emsp; v0.4.8 &emsp; 修复站点截图模块一些边界条件问题  
2024-01-16 &emsp; v0.4.9 &emsp; 内部HTTP API升级改造  
2024-01-16 &emsp; v0.4.9 &emsp; 内部HTTP API支持展示上一次历史任务信息  
2024-01-17 &emsp; v0.4.9 &emsp; 内部HTTP API任务详细信息数据结构优化  
2024-01-17 &emsp; v0.4.9 &emsp; 扫描任务状态相关功能优化  
2024-01-17 &emsp; v0.4.9 &emsp; 支持展示和动态计算扫描开始时间,扫描结束时间和扫描耗时  
2024-01-17 &emsp; v0.4.9 &emsp; 实时上报扫描结果时, 加入了扫描任务ID, 方便后期查询           
2024-01-17 &emsp; v0.4.9 &emsp; 整理完善文档          
2024-01-29 &emsp; v0.4.10&emsp; 修复上报的资产信息不带location的问题  
2024-01-29 &emsp; v0.5.0 &emsp; 新增扫描手动暂停接口   
2024-03-02 &emsp; v0.5.0 &emsp; 优化内部任务调度API, 修复若干问题   
2024-03-04 &emsp; v0.5.1 &emsp; 尝试解决超长时间任务无法退出问题   
2024-03-05 &emsp; v0.5.2 &emsp; 现针对扫描目标输入,支持多个IP或IP段组合(beta)  
2024-03-06 &emsp; v0.5.3 &emsp; 处理参数设置互相关联的条件下, 特定一些设置会触发crash问题  
2024-03-06 &emsp; v0.5.3 &emsp; 新增自定义是否上报结果的参数  







## 功能
1. 主机存活扫描(icmp)
2. 主机端口开放扫描(支持syn和tcp全连接方式)
3. 主机端口服务识别
4. HTTP指纹识别
5. 数据上报(基于lua接口的管道)
6. 实现内部API远程调用的接口(创建扫描与获取实时扫描状态)
7. 支持CIDR和地址范围(192.168.1.1-100)格式的目标资产输入
8. 支持禁止扫描的白名单ip设置, 支持ip,CIDR,地址范围以及用","分割组合输入
9. web站点截图功能, 并上传至minio图床
10. 支持三方指纹库动态加载和更新(基于三方依赖)
11. 设置扫描时间段(定时暂停与开始)


## todo
1. 遇到一些边界条件时稳定性优化
2. syn扫描时实时显示进度 
3. 常见UDP协议扫描
4. 优化扫描速度
5. web HTTP指纹识别优化
6. web HTTP指纹数据库自定义(中心端)
7. TCP 指纹模块优化(支持更多协议)
8.  TCP 指纹模块加入安全相关指纹(FRP,CS listener,msf listener....)
9.  优化扫描结果的数据结构
10. 分布式集群扫描,智能分配扫描任务
11. 处理模块实时返回数据的问题  
12. 支持端口排除项(端口,端口范围以及用","分割组合输入)
13. 尝试处理 lua脚本更新 旧的扫描任务没有强制停止问题  
14. 启发式扫描, 针对超大网段的扫描效率优化    
15. ……  

## Lua API
  直接查看示例  

## 内部HTTP API
### **GET** `/api/v1/arr/agent/radar/status`  
获取当前扫描服务状态   
### **POST** `/api/v1/arr/agent/radar/runscan`  
运行扫描任务(如果已有扫描任务正在进行则无法运行)  
**参数**   ( * 为必填项):  
`target`  *  目标IP/CIDR/IP范围  
`location`  *  网络位置  
`name`  *  任务名称  
`mode`  模式 "tcp"(默认)/"syn"  
`port`  端口  默认top1000  
`rate`  传输层协议基础发包速率   
`timeout`  超时时间(ms)  
`httpx`  是否开启http指纹探测  
`fingerDB`  指定web指纹库三方依赖(不指定则使用内置默认指纹库)   
`screenshot`  是否开启站点截图功能  
`ping`  是否开启ping存活探测  
`pool_ping`  ping探测协程数      
`pool_scan`  scan协程数  
`pool_finger` 指纹识别协程数   
`excludeTimeRange`  扫描排除时间段 示例"daily,9:00,17:00"  

**例子**:  
```json
{
    "target":"192.168.1.1/24",
    "location":"测试本地网",
    "name":"测试扫描任务",
    "port":"top1000",
    "httpx":true,
    "screenshot":true
}
```




## 示例
```lua
local rr = vela.radar{
  name = "radar",
  finger = {timeout = 500 , udp = false , fast = false},
  minio = {accessKey="xxx" , secretKey="xxx" , endpoint="xxx" , useSSL=false}
}

local es = vela.elastic.default("vela-radar-%s" , "$day")
rr.pipe(function(host)
  es.send(host)
end)

-- 启动服务
rr.start()

-- 启动内部API调用功能
rr.define()

-- 开启扫描任务
rr.task("192.168.1.1/24").port("top1000").httpx(true).exclude("192.168.1.100,192.168.1.10-20").run()
-- 关闭主机ping存活探测 .ping(false)
-- web快照截图 .screenshot(true)
-- 指定指纹库  .fingerDB("radar-http-finger.json")
-- 指定扫描时间段  .excludeTimeRange("daily","15:00","15:02")
```

## 注意
1. 在做外网探测时, 可能会因为syn的包过多, 导致网络无法链接