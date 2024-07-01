# TbSign➡️ (Dev)

---

只是一个签到程序，可以配合[百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)使用，也可以独立运作

## ⚠ 警告

- 随缘维护，不保证不会封号，请**不要**用于主力账号！！！
- 不接受任何路线图以外的 PR，路线图内的 PR 不一定会接受

## flags

| flag     | default          | description                        |
| :------- | :--------------- | :--------------------------------- |
| username |                  | 数据库帐号                         |
| pwd      |                  | 数据库密码                         |
| endpoint | `127.0.0.1:3306` | 数据库端点                         |
| db       | `tbsign`         | 数据库名称                         |
| db_path  |                  | SQLite 文件目录                    |
| test     | `false`          | 测试模式，此模式下不会运行计划任务 |
| api      | `false`          | 是否启动后端 api                   |
| address  | `:1323`          | 后端运行地址                       |
| setup    | `false`          | 安装程序                           |

示例

```shell
go run main.go --username=<dbUsername> --pwd=<DBPassword>
# or
./tbsign_go --username=<dbUsername> --pwd=<DBPassword>
# or https://github.com/cosmtrek/air
air -- --db_path=tbsign.db --test=true --api=true
```

## env

不支持 `.env` 文件，请直接设置环境变量，使用顺序是 `flags` > `env` > `default`

| flag        | description                        |
| :---------- | :--------------------------------- |
| tc_username | 数据库帐号                         |
| tc_pwd      | 数据库密码                         |
| tc_endpoint | 数据库端点                         |
| tc_db       | 数据库名称                         |
| tc_db_path  | SQLite 文件目录                    |
| tc_test     | 测试模式，此模式下不会运行计划任务 |
| tc_api      | 是否启动后端 api                   |
| tc_address  | 后端运行地址                       |

## 数据库

只要 `db_path`/`tc_db_path` 的值不为空字符串，就会使用 SQLite；否则使用 MySQL

## 安装

- 迁移法 (仅 MySQL)
  - 直接安装 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)
  - 使用原数据库配置启动
- 全新安装 (MySQL, SQLite)
  - 启动程序，添加 `flag` `setup=true` (示例: `./tbsign_go --db=tbsign.db --api=true --address=:8080 --setup=true`)
  - 根据文字提示完成自动安装流程

    ```plaintext
    ➜  tbsign_go git:(master) ✗ ./tbsign_go --db_path=tbsign1.db --api=true --setup=true
    2024/06/29 18:18:47 tbsign: sqlite connected!
    📌现在正在安装 TbSign➡️，如果数据库内含有数据，这样做会导致数据丢失，请提前做好备份。
    如果已经完成备份，请输入以下随机数字并按下回车（显示为 "--> 1234 <--" 代表需要输入 "1234"）
    --> 428650691 <--
    请输入: 428650691
    ⌛正在清理旧表
    ⌛正在建立数据表和索引
    ⌛正在导入数据...
    ⌛导入第1项...
    ⌛导入第2项...
    🔒注册管理员帐号...
    管理员用户名: a
    管理员邮箱: a@a.a
    管理员密码 (注意空格): a
    ⌛正在注册管理员帐号...
    🎉安装成功！请移除掉 `--setup=true` 后重新执行本文件以启动系统
    ➜  go_tbsign git:(master) ✗ ./tbsign_go --db_path=tbsign1.db --api=true
    ```

  - 去掉 `--setup=true`，启动程序
  - \*(选做) 登录管理员帐号，打开 **系统管理** 即可手动开启自带插件

### 兼容性

- [⚠️] 通过 *迁移法* 启动的程序能够与 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/) 共存，但无法使用 *用户成长任务* 插件，因为这是一个未公开的插件
- [❌] 通过 *全新安装* 启动的程序无法兼容 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)，因为缺少相关数据表和设置选项

## 前端

➡️ <https://github.com/BANKA2017/tbsign_go_fe>

## API (WIP)

仅供参考，未来可能还会修改，等到稳定后随缘出文档

## 插件

部分功能来自原版官方插件

- [x] [自动刷新贴吧列表](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ref)
- [x] [名人堂](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_rank)
- [x] [循环封禁](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ban)
- [x] ?用户成长任务(beta)

## 已知问题

- [x] 不支持限制单次签到贴吧总数，会一次性全部签完
- [x] 没有灵活的请求间隔和请求头模拟，有封号的风险
- [x] 只有签到这一个功能，对重签的处理约等于没有
- [ ] 混乱的输出，随处可见的打点
- [ ] 没有进行任何优化
- [x] 一次性执行，仍然需要 cron
- [x] 会运行未安装/未激活的插件
- [ ] 邀请码系统在考虑要不要做，如果不想做就会移入**不会解决列表**
- [x] 邮箱相关……邮箱找回✅

**下面提到的不会解决**

- 兼容性
  - [x] 不会对所谓的 vip 帐号有任何特殊照顾，原 vip 帐号的特权也会下放给所有账号
  - [x] 无法支持一个贴吧账号绑定到多个云签帐号
  - [x] 不同语言各有特性，不会强求 1:1 兼容
  - [x] 不支持自定义的数据表前缀，统一使用默认前缀 `tc_`
  - [x] 不支持分表
  - [x] 登录/注册不会有验证码
- 接口
  - [x] 循环封禁无法确认封禁是否成功，因为返回的结果是一样的
  - [x] 执行成长任务时无法确认是否重复执行，原因同上

## TODO

\* 带 `?` 开头的表示有考虑，但可能永远不做

- [ ] 解决已知问题
- [ ] 兼容官方已收录插件中关于贴吧的部分
  - [x] [自动刷新贴吧列表](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ref)
  - [x] [名人堂](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_rank)
  - [x] [循环封禁](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ban)
  - [ ] ?[删贴机](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_review)（可能会拖很久甚至不会做）
- [ ] 优化 PHP 原版的相关功能
- [x] 不再考虑 ~~通过读取 `config.php` 取得数据库连接信息~~
- [x] 自动清理解除绑定的帐号的插件设置
- [ ] ?打包/Docker/或者别的
- [ ] ?自动化部署
- [ ] ?支持更多 Gorm 也支持的数据库
- [ ] ?邮箱以外的推送方式
- [ ] 完善权限控制
- [ ] 个人数据导出 (接口 `/passport/export` 已写好，但没想好如何处理好安全问题，当前所有接口都会自动删除 `bduss` 和 `stoken` 的值，但导出会不可避免地需要处理这个问题)
- [ ] ……更多的想起来再加

## 感谢

- [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)
- [Starry-OvO/aiotieba](https://github.com/Starry-OvO/aiotieba/)
