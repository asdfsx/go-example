# 海康威视

## 启用onvif的流程

* 登录海康威视摄像头的管理页面

* 进入配置->高级配置->集成协议

* 启用ONVIF

* 创建账号 operator、video，用户类型分别是操作员、视频用户

注意如果用户类型是管理员的话，无法用onvif登录，
即使使用官方的`onvif device test tool`也无法登录

## onvif 云台移动 参数对照

ONVIF参数取值	OSD显示度数
-1（最小）	0
-0.5	90
0	180
1（最大）	359
ONVIF Absolute move - Tilt

| ONVIF参数取值 | OSD显示度数 |
| --- | --- |
| -1（最小）| 0（水平）|
| -0.5 | 22.5 |
| 0 | 45 |
| 0.5 |	67.5 |
| 1（最大）|	90（垂直向下）|

## 问题

目前使用的包：

github.com/yakovlevdmv/goonvif 停止维护。导致其他fork出来的版本，解决了各自发现的bug后，都没有合并到一起。

github.com/rikugun/goonvif 添加了一些自己的一些改动，但是还是需要用原版的很多东西。

所以需要同时使用两个goonvif（恶心）
