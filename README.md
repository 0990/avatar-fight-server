# avatar-fight-server
头像大乱战golang服务端  
点此试玩：http://34.92.209.203

## 服务器结构
基于github.com/0990/goserver服务器框架的游戏  
有以下服务构成  
1,gate服，负责消息转发  
2,game服，负责游戏逻辑  
3,center服,中心服，负责玩家管理  

## 编译运行
1，先运行nats-streaming-server(https://github.com/nats-io/nats-streaming-server)  
2，执行build.bat 会生成bin/avatar-fight-server.exe,运行即可 
 
客户端代码在此：https://github.com/0990/avatar-fight-client

## 说明
1,服务器架构上支持多game服，目前未完善此部分  
2,后续会加上对微信小游戏账号登录的支持