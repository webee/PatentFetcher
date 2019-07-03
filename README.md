# PatentFetcher
中国专利公告爬虫后端，分配任务和收集结果

前端：https://greasyfork.org/zh-CN/scripts/387043-cnpatentfetcherclient

## 说明
朋友让我帮忙爬取一些数据，由于目标网站采用的防爬虫策略，简单的模拟请求不行，使用headless browser简单试了一下，也不行，
于是想到在前端使用js爬取再提交到后端，在浏览器中简单实验之后发现是可行的。

后端提供两个接口给前端爬虫使用：分配任务，提交结果
打开多个标签(注：浏览器有单个ip连接数限制)，或多个浏览器, 甚至开不同的机器就可以实现分布式并行前端爬取了。


## 设想
目前只是解决这一个网站的爬取问题，但我想这种方法是可以通用的，也是可行的。  
前端爬取其实跟使用headless browser（Selenium，PhantomJS等）原理一样的，但这里的关键是分散地使用真实的客户端,
利用人们空闲的计算和流量资源，而不是使用自己的服务器。这可以解决许多复杂爬取面临地很多痛点：ip限制，人机验证，足够的服务器等
是否可以设计通用的后端和客户端，给用户付费爬取或者开发者之前积分互助爬取，以解决一些复杂网站的爬取问题。
