## https://github.com/Chanzhaoyu/chatgpt-web 的golang 后端实现

> 本项目是基于 [chatgpt-web](https://github.com/Chanzhaoyu/chatgpt-web) 的后端实现，前端实现请移步 [chatgpt-web](https://github.com/Chanzhaoyu/chatgpt-web) 。
> 兼容原版接口


* 实现了key池，可以将有效key放在根目录enable.txt目录下，key消耗完之后，会自动从key池中获取新的key
* 实现了自定义模型 温度
* 实现了用户可以传递自己的key


## 使用方法

* 部署 [chatgpt-web](https://github.com/Chanzhaoyu/chatgpt-web) 项目
* 部署本项目
* 将本项目的地址填入 [chatgpt-web](https://github.com/Chanzhaoyu/chatgpt-web）的.env中的VITE_APP_API_BASE_URL中


## 部署本项目的方法

* 从 [release](#) 中下载最新的版本 
* 将下载的文件解压到任意目录
* 将enable.txt中的key替换为自己的key
* 运行chatgpt 
