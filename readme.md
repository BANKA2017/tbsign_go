# Tieba-Cloud-Sign-Go (Dev)
---

只是一个签到程序，需要配合[百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)使用

## ⚠ 警告

不保证持续维护，不保证不会封号，请**不要**用于主力账号！！！

### flags

| flag     | default          | description |
| :------- | :--------------- | :---------- |
| account  |                  | 数据库帐号  |
| pwd      |                  | 数据库密码  |
| endpoint | `127.0.0.1:3306` | 数据库端点  |
| tbsign   | `tbsign`         | 数据库名称  |


示例

```shell
go run run.go --account=<DBAccount> --pwd=<DBPassword>
#or
./run  --account=<DBAccount> --pwd=<DBPassword>
```

### 已知问题

- [ ] 不支持自定义的数据表前缀，统一使用默认前缀 `tc_`
- [ ] 不支持分表，暂时也没有支持的打算
- [ ] 不支持限制单次签到贴吧总数，会一次性全部签完
- [ ] 没有灵活的请求间隔和请求头模拟，有封号的风险
- [ ] 只有签到这一个功能，对重签的处理约等于没有
- [ ] 混乱的输出，随处可见的打点
- [ ] 没有进行任何优化
- [ ] 一次性执行，仍然需要 cron

### TODO

- [ ] 解决已知问题
- [ ] 兼容官方已收录插件中关于贴吧的部分
  - [ ] 自动刷新贴吧列表
  - [ ] 名人堂
  - [ ] 循环封禁
  - [ ] 删贴机（可能会拖很久甚至不会做）
- [ ] 反向优化 PHP 原版的相关功能
- [ ] 通过读取 `config.php` 取得数据库连接信息
- [ ] ……更多的想起来再加
