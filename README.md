# vela-radar
Active scanning of network assets  
利用ssoc agent实现网络资产主动扫描  


## 更新日志
2023-11-10 &emsp; v0.1.0 &emsp; 初始测试版本  
2023-11-13 &emsp; v0.1.0 &emsp; 资产扫描功能优化, 加入web指纹探测功能  
2023-11-15 &emsp; v0.1.0 &emsp; 优化扫描数据的处理, 对接后端上报接口的数据结构  
2023-11-17 &emsp; v0.1.0 &emsp; 实现远程调用的内部API, 实时进度显示  


## 功能
1. 主机存活扫描(icmp)
2. 主机端口开放扫描(支持syn和tcp全连接方式)
3. 主机端口服务识别
4. HTTP指纹识别
5. 数据上报(基于lua接口的管道)
6. 实现内部API远程调用的接口(创建扫描与获取实时扫描状态)
7. 支持CIDR和地址范围(192.168.1.1-100)格式的目标资产输入

## todo
1. 黑名单IP
2. 设置扫描时间段(定时暂停与开始)
3. 遇到一些边界条件时稳定性优化
4. syn扫描时实时显示进度 
5. 常见UDP协议扫描
6. 优化扫描速度
7. web HTTP指纹识别优化
8. web HTTP指纹数据库自定义(中心端)
9. TCP 指纹模块优化(支持更多协议)
10. TCP 指纹模块加入安全相关指纹(FRP,CS listener,msf listener....)
11. 优化扫描结果的数据结构
12. 分布式集群扫描,智能分配扫描任务
13. 处理模块实时返回数据的问题  
14. ……  

## 内部API
### **GET** `/arr/agent/radar/status`  
获取当前扫描服务状态   
### **POST** `/arr/agent/radar/runscan`  
运行扫描任务(如果已有扫描任务正在进行则无法运行)  
**参数**   ( * 为必填项):  
`target`  *  目标IP/CIDR/IP范围  
`location`  *  网络位置  
`name`  *  任务名称  
`mode`  模式 "tcp"(默认)/"syn"  
`port`  端口   
`rate`  基础发包速率   
`timeout`  超时时间(ms)  
`httpx`  是否http指纹探测  
`ping`  是否开启ping存活探测  
`pool_ping`  ping探测协程数      
`pool_scan`  scan协程数  
`pool_finger` 指纹识别协程数   

**例子**:  
```json
{
    "target":"192.168.1.1/24",
    "location":"测试本地网",
    "name":"测试扫描任务",
    "port":"top1000",
    "httox":true
}
```




## 示例
```lua
local rr = vela.radar{
  name = "radar",
  finger = {timeout = 500 , udp = false , fast = false},
}

local es = vela.elastic.default("vela-radar-%s" , "$day")
rr.pipe(function(host)
  es.send(host)
end)

-- 外部API
rr.define()

-- 启动服务
rr.start()

-- 启动内部API调用功能
rr.define()

-- 开启扫描任务
rr.task("127.0.0.1").port("top1000").httpx(true).run()
```

## 注意
1. 在做外网探测时, 可能会因为syn的包过多,导致网络无法链接