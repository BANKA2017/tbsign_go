# Tieba-Cloud-Sign-Go (Dev)

---

只是一个签到程序，需要配合[百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)使用

## ⚠ 警告

- 不保证持续维护，不保证不会封号，请**不要**用于主力账号！！！
- 没有文档，没有教程，不会回答任何疑问

### flags

| flag     | default          | description |
| :------- | :--------------- | :---------- |
| username |                  | 数据库帐号  |
| pwd      |                  | 数据库密码  |
| endpoint | `127.0.0.1:3306` | 数据库端点  |
| db       | `tbsign`         | 数据库名称  |

示例

```shell
go run run.go --username=<dbUsername> --pwd=<DBPassword>
#or
./run --username=<dbUsername> --pwd=<DBPassword>
```

### env

不支持 `.env` 文件，请直接设置环境变量，使用顺序是 `flags` > `env` > `default`

| flag        | description |
| :---------- | :---------- |
| tc_username | 数据库帐号  |
| tc_pwd      | 数据库密码  |
| tc_endpoint | 数据库端点  |
| tc_db       | 数据库名称  |

### 已知问题

- [ ] 不支持自定义的数据表前缀，统一使用默认前缀 `tc_`
- [ ] 不支持分表，暂时也没有支持的打算
- [ ] 不支持限制单次签到贴吧总数，会一次性全部签完
- [ ] 没有灵活的请求间隔和请求头模拟，有封号的风险
- [ ] 只有签到这一个功能，对重签的处理约等于没有
- [ ] 混乱的输出，随处可见的打点
- [ ] 没有进行任何优化
- [ ] 一次性执行，仍然需要 cron

**下面提到的不会解决**

- [x] 不会对所谓的 vip 帐号有任何特殊照顾，原 vip 帐号的特权也会下放给所有账号
- [x] 无法支持一个贴吧账号绑定到多个云签帐号
- [x] 不同语言各有特性，不会强求 1:1 兼容
- [x] 循环封禁无法确认封禁是否成功，因为返回的结果是一样的

### TODO

- [ ] 解决已知问题
- [ ] 兼容官方已收录插件中关于贴吧的部分
  - [x] [自动刷新贴吧列表](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ref)
  - [x] [名人堂](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_rank)
  - [x] [循环封禁](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ban)
  - [ ] [删贴机](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_review)（可能会拖很久甚至不会做）
- [ ] 优化 PHP 原版的相关功能
- [x] 不再考虑 ~~通过读取 `config.php` 取得数据库连接信息~~
- [ ] 自动清理解除绑定的帐号的插件设置
- [ ] ……更多的想起来再加
