# webook

A simple web application back-end writing in Go. Using MySQL as the database.

Now supports:
+ Sign up new user.
+ Log in existing user.
+ Edit user's profile.
+ Query user's profile.

A lot of new features on the way...

---

#### JWT

- 组合 jwt.RegisteredClaims 来实现 Claims 接口
- 适合在分布式环节中使用
- 不依赖第三方存储
- 性能较好，没有访问瓶颈
- 安全性比较依赖加密算法
- token 会传到前端，不要放置敏感信息
- 优雅退出登陆，使 token 失效方法？
  1. 布尔过滤器
  2. redis 存一个黑名单
- JWT 和 session 混合使用机制：敏感数据存在 session 中，使用 JWT 中的 userId 组成访问 session（使用redis存储）的 key 来进行访问
- ExpiresAt 设置过期时间

---

#### lecture 8

- 系统保护：限流
  1. 如何识别限流对象：IP（存在共享IP的情况）
  2. 限流阈值设置：通过压测整个系统来得到
- 用Redis限流：考虑多个实例部署，经过负载均衡之后，统计单一实例的请求数没有意义，需要Redis来统计所有实例的请求总数。
- 增强登陆安全：
  1. 登录的额外信息（在JWT中加入浏览器的User-Agent头部，在中间件中进行校验）
  2. IP归属地（不能使用IP，尤其是移动网络）

---

#### week3 homework

![](/Users/cheng/go/src/webook/image/kubectl command image.png)



