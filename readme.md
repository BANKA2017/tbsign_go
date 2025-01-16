# TbSign➡️

---

只是一个签到程序，可以配合[百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)使用，也可以独立运作

## ⚠ 注意

- 随缘维护，不保证不会封号，请**不要**用于主力账号！！！
- 不接受任何路线图以外的 PR，路线图内的 PR 不一定会接受

## flags

| flag           | default          | description                                                                   |
| :------------- | :--------------- | :---------------------------------------------------------------------------- |
| username       |                  | 数据库账号                                                                    |
| pwd            |                  | 数据库密码                                                                    |
| ~~endpoint~~   | `127.0.0.1:3306` | 数据库 `host:port` （已废弃）                                                 |
| host           | `127.0.0.1:3306` | 数据库 `host:port`                                                            |
| db             | `tbsign`         | 数据库名称                                                                    |
| db_tls         | `false`          | CA 证书的选项（仅用于 MySQL），请看 [CA 证书](#ca-证书) 部分                  |
| db_path        |                  | SQLite 文件目录                                                               |
| test           | `false`          | 测试模式，此模式下不会运行计划任务                                            |
| api            | `false`          | 是否启动 api                                                                  |
| fe             | `false`          | 是否启动自带前端，仅当 `api` 为 `true` 时可为 `true`                          |
| address        | `:1323`          | 后端运行地址                                                                  |
| setup          | `false`          | 强制安装程序（可能会覆盖现有配置）                                            |
| auto_install   | `false`          | 自动安装（仅当数据库不存在时安装）                                            |
| admin_name     |                  | 管理员账号，仅当 `auto_install` 为 `true` 时会用到                            |
| admin_email    |                  | 管理员邮箱，仅当 `auto_install` 为 `true` 时会用到                            |
| admin_password |                  | 管理员密码，仅当 `auto_install` 为 `true` 时会用到                            |
| no_proxy       | `false`          | 忽略环境变量中的代理配置                                                      |
| allow_backup   | `false`          | 允许用户批量导出/导入账号和贴吧列表，建议阅读 readme.md 的 [备份](#备份) 部分 |

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

| flag              | description                                                                   |
| :---------------- | :---------------------------------------------------------------------------- |
| tc_username       | 数据库账号                                                                    |
| tc_pwd            | 数据库密码                                                                    |
| ~~tc_endpoint~~   | 数据库 `host:port` （已废弃）                                                 |
| tc_host           | 数据库 `host:port`                                                            |
| tc_db             | 数据库名称                                                                    |
| tc_db_tls         | CA 证书的选项（仅用于 MySQL），请看 [CA 证书](#ca-证书) 部分                  |
| tc_db_path        | SQLite 文件目录                                                               |
| tc_test           | 测试模式，此模式下不会运行计划任务                                            |
| tc_api            | 是否启动后端 api                                                              |
| tc_fe             | 是否启动自带前端，仅当 `tc_api` 为 `true` 时可为 `true`                       |
| tc_address        | 后端运行地址                                                                  |
| tc_auto_install   | 自动安装（仅当数据库不存在时安装）                                            |
| tc_admin_name     | 管理员账号，仅当 `tc_auto_install` 为 `true` 时会用到                         |
| tc_admin_email    | 管理员邮箱，仅当 `tc_auto_install` 为 `true` 时会用到                         |
| tc_admin_password | 管理员密码，仅当 `tc_auto_install` 为 `true` 时会用到                         |
| tc_allow_backup   | 允许用户批量导出/导入账号和贴吧列表，建议阅读 readme.md 的 [备份](#备份) 部分 |

## 数据库

只要 `db_path`/`tc_db_path` 的值不为空字符串，就会使用 SQLite；否则使用 MySQL

MySQL 要求支持 [MySQL 窗口函数](https://dev.mysql.com/doc/refman/8.0/en/window-functions.html)

- MySQL >= [8.0](https://dev.mysql.com/doc/refman/8.0/en/window-functions.html)
- MariaDB >= [10.2](https://mariadb.com/kb/en/changes-improvements-in-mariadb-10-2/)
- TiDB >= [3.0](https://docs.pingcap.com/tidb/stable/release-3.0-ga)

### CA 证书

有的云服务，要求用户使用 TLS 连接到它们的数据库；有的用户对内网环境有特殊的安全需求，此时可能需要用到证书文件

可用于 `db_tls` 的值包括 `true`, `false`, `skip-verify`, `preferred` 以及证书文件的路径，更多信息请参考 [go-sql-driver/mysql#tls](https://github.com/go-sql-driver/mysql?tab=readme-ov-file#tls) 

例如下面的第二项是 Debian/Ubuntu 的目录；如果证书尚未被导入到系统，系统管理员可能就需要这样填写证书的实际目录

```shell
go run main.go --db_tls=true
go run main.go --db_tls=/etc/ssl/certs/ca-certificates.crt
```

## 安装

- 迁移法 (仅 MySQL)
  - 直接安装 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)
  - 使用原数据库配置启动
- 全新安装 (MySQL, SQLite)
  - 手动安装
    - 启动程序，添加 `flag` `setup=true` (示例: `./tbsign_go --db_path=tbsign.db --api=true --address=:8080 --setup=true`)
    - 根据文字提示完成自动安装流程（不同情况下文字提示可能会略有不同）

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
      🔒注册管理员账号...
      管理员用户名: a
      管理员邮箱: a@a.a
      管理员密码 (注意空格): a
      ⌛正在注册管理员账号...
      🎉安装成功！
      ```

    - \*(选做) 登录管理员账号，打开 **系统管理** 即可手动开启自带插件
  - 自动安装
    - 启动程序，添加 `flags` 或 环境变量 (示例: `./tbsign_go --db=tbsign --username tcdb --pwd tcdb_password --api=true --address=:8080 --auto_install=true --admin_name=a --admin_email=a@a.a --admin_password=a`)

      - 担心 log 泄露信息的此时可以设随机值，等到安装完成后再登录修改
      - 除非数据库被删除，否则 用户名/邮箱/密码 仅在首次开启时会用到
    - \*(选做) 登录后台修改信息，开启自带插件

\* 注：`setup` 和 `auto_install` 不可同时为 `true`

### 兼容性

- [⚠️] 通过 *迁移法* 启动的程序能够与 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/) 共存，但部分插件没有对应的 PHP 版本
- [❌] 通过 *全新安装* 启动的程序无法兼容 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)，因为缺少部分数据表和设置选项
- [❌] MySQL 不支持无法使用 [MySQL 窗口函数](https://dev.mysql.com/doc/refman/8.0/en/window-functions.html) 的发行版，使用了**不支持**的函数 `ROW_NUMBER()`
- [⚠️] 开发环境使用的 MySQL 版本号为 `8.0` 或更高，其他 MySQL 发行版请自行测试是否可用
- [⚠️] xgo 镜像使用的 tag 为 `go-1.23.1`，可能会有操作系统不再受到支持（如 Windows 7）

### Docker

因为各种原因想要使用 Docker 部署的用户，请参考下面的 compose 配置，官方 releases 支持的 `os-arch` 组合是 `linux-amd64` 和 `linux-arm64`

```yml
# docker-compose.yml
# docker compose up -d
services:
  tbsign-go:
    image: alpine:3                           # 如果 alpine:3 无法运行，可以改为 debian:12
    container_name: tbsign-go
    environment:
      tc_db_path: "/app/tbsign/tbsign_go.db"  # SQLite 目录，填的是容器的 path，而不是宿主机的 path，建议跟可执行文件放在一起
      tc_api: "true"
      tc_fe: "true"
      # ...
      # 其他环境变量请参考前面 #env 部分
      ## 可能只支持自动化安装，或者提前安装好再迁移数据库
    volumes:
      - /path/to/binary/file:/app/tbsign      # 将 `/path/to/binary/file` 改成可执行文件所在的目录（不含可执行文件的名字）
    working_dir: /app/tbsign
    command: ["sh", "-c", "./tbsign_go"]      # 如果文件名不叫 tbsign_go，别忘了改名；如果使用 debian:12，要改为 ["sh", "-c", "apt update && apt install -y ca-certificates && update-ca-certificates && ./tbsign_go"]
    ports:
      - 8080:1323                             # 宿主机端口:容器内端口，根据实际情况修改
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: 50m
        max-file: "3"
```

## 前端

➡️ <https://github.com/BANKA2017/tbsign_go_fe>

### 嵌入式前端

此类部署需要用到其他仓库，并且需要手动或自动添加文件，如果不知道是怎么回事，请忽略本节，使用前后分离部署

- 使用嵌入式前端前，请设置环境变量 

  ```env
  NUXT_BASE_PATH="/api"
  ```

- 在前端仓库的目录执行 `yarn run generate` 后将 `/dist` 目录整个拷贝到 `/assets` 内（即 `/assets/dist`）
- 照常 `go build`

嵌入式前端启用后，API 路径自动添加前缀 `/api`，同时 response header 不再含有 `Access-Control-Allow-Origin`

## API (WIP)

仅供参考，未来可能还会修改，等到稳定后随缘出文档

## 插件

外链为对应 PHP 版插件

- [x] [自动刷新贴吧列表](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ref)
- [x] [名人堂](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_rank)
- [x] [循环封禁](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ban)
- [x] 用户成长任务
- [x] [知道商城抽奖](https://github.com/96dl/Tieba-Cloud-Sign-Plugins/blob/master/ver4_lottery/ver4_lottery_desc.php)
- [x] ⚠️文库任务，*使用此插件可能会导致文库签到功能被封禁，封禁理由为：**您的账号因涉嫌刷分作弊而被封禁，不能进行此项操作***
- [x] 吧主考核

### 插件开发

⚠ 插件还无法做到开箱即用，使用前需要通过源码编译

目前插件开发至少需要分别准备一个前端(.vue)和后端的文件(.go)，配置方面可能会比较麻烦

- 插件目录 `/plugins` 放下核心文件，这些前缀只是为了便于管理，golang 中同一层的非隐藏文件命名对编译无影响
  - 标准插件
    - 文件命名应当使用前缀 `s_`(standard)，插件本身的命名建议使用 `<namespace>_<plugin_name>` 来避免冲突，但 `namespace` 为非强制项，实际命名请尽量贴近插件功能
      - 命名建议使用 **英语单词/常用缩写**，或者与 **完整拼音** 的组合。请不要用拼音首字母缩写
    - 导出 API 接口的方式请参考现有的插件
      - 导出路由的路径为 `[/api]/plugins/<PluginInfo.Name>/<PluginInfo.Endpoints.Path>`，有没有 `/api` 前缀取决于是否启用自带前端
    - 其他步骤请参考几个自带的标准插件，必要的函数和变量配置请参考 [标准插件模板](https://github.com/BANKA2017/tbsign_go/blob/master/plugins/_s_example.go)
  - 升级插件
    - 文件命名应当使用前缀 `u_`(upgrade)，用于不兼容更新，但目前还不存在需要用到的场景，所以暂时无需理会（甚至根本没有写调用逻辑）
  - 核心插件
    - 文件命名应当使用前缀 `c_`(core)
- 前端
  - 对应插件的页面
    - 在目录 `pages/` 添加对应插件的页面 vue 文件，文件名必须带有前缀 `plugin_`
  - *如果对变量设置表单有特殊要求，还需要修改 `system_admin.vue` 的内容
    - 含有特殊后缀 `_action_limit` 的变量将会在前端自动识别成不小于 `0` 的 `number` 类型
    - 其他变量识别成 `text`
- *数据库 model
  - 简单的信息可以直接存在表 `tc_options`（全局） 或 `tc_users_options`（用户） 中
  - 如果需要使用数据库存储插件信息
    - 请准备好表模型并存放在 `model/`，包名为 `model`
    - 或者生成好表模型后将相关内容拷贝到插件的 go 文件内，包名为 `_plugin`

    参考 codegen 命令为

    ```shell
    /root/go/bin/gentool -dsn "..." -modelPkgName "model" -onlyModel
    ```

    尽管 Gorm 可以手搓 raw SQL，但我们并不建议这样做

- *protobuf
  - 部分请求可能需要用到 protobuf，这部分还没完善，敬请期待

标准插件不再与核心强绑定，使用 [SemVer](https://semver.org/lang/zh-CN/)，但新插件使用的一些新版核心函数可能会导致无法在旧版核心中使用

*关于插件开发配置上越来越复杂的吐槽：为了逃离配置地狱而亲手创造了一个配置地狱，也是很讽刺了

## 编译

### CGO

> go-sqlite3 is cgo package. If you want to build your app using go-sqlite3, you need gcc. However, after you have built and installed go-sqlite3 with `go install github.com/mattn/go-sqlite3` (which requires gcc), you can build your app without relying on gcc in future.

### build.sh

简单写了个编译脚本，存放在 `build.sh`

默认 `CGO_ENABLED=0`，如果编译结果无法运行可以尝试改为 `CGO_ENABLED=1`

### 版本号

格式为 `tbsign_go.<YYYYMMDD>.<BACKEND_COMMIT_HASH[0:7]>.<FRONTEND_COMMIT_HASH[0:7]>.<OS>-<ARCH>` ，如果是用于 Windows 系统的二进制还会有 `.exe` 拓展名

官方提供的二进制包含 linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64 五个版本，使用 [xgo](https://github.com/techknowlogick/xgo) 进行交叉编译，手动触发编译的 Actions 任务

其它系统请自行编译运行

## 备份

从 [tbsign_go.20241203.f7b5434.881b23b](https://github.com/BANKA2017/tbsign_go/tree/tbsign_go.20241203.f7b5434.881b23b) 版本起，支持用户导出/导入部分账号数据。

此前，tbsign_go 设计上不允许 API 返回 `bduss` 和 `stoken`，只能通过查看数据库获取，确保即使站点账号被盗，仍无法从 tbsign_go 导出敏感数据。此版本后，若站点开放备份功能，用户登录后可通过账号密码导出这些数据。

为平衡**安全**与**便利**，系统设置页面（网页）新增了备份功能开关，并添加了启动时确定是否启用的总开关（cli）。未启用的站点将禁用相关 API，是否开启功能由站点管理员自行决定。

### 格式

导出格式与对应数据表一致，可通过编辑文件实现批量导入账号

```json
{
  "tc_baiduid": [
    {
      "id": 1,
      "bduss": "",
      "stoken": "",
      "portrait": "tb.1.xxx.xxxx",

      "name": ""
    }
  ],
  "tc_tieba": [
    {
      "pid": 1,
      "fid": 29,
      "tieba": "小说",

      "no": false,
      "status": 0,
      "latest": 3,
      "last_error": "NULL"
    }
  ]
}
```

- `tc_baiduid.id` 的值未必等于最终该账号的 `pid`，此值仅用于映射 `tc_tieba.pid`
- `tc_tieba.pid` 必须要有对应的 `tc_baiduid.id` 值
- `tc_baiduid.name`、`tc_tieba.no`、`tc_tieba.status`、`tc_tieba.latest`、`tc_tieba.last_error` 可以没有值，但建议加上
  - `name` 是用户名，有些后期通过手机注册的用户可能没有用户名
  - `no` 用于表示是否**忽略**签到
  - `status` 和 `last_error` 用于表示上一次签到的情况
  - `latest` 表示最后签到的日期，每月第一天是 `1`

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
- [x] ~~签发用户 JWT 的密钥仅存在于内存中，系统重启后会掉登录~~ ~~特性！都是特性！~~ 不再使用 JWT

**下面提到的不会解决**

- 兼容性
  - [x] 不会对所谓的 vip 账号有任何特殊照顾，原 vip 账号的特权也会下放给所有账号
  - [x] 无法支持一个贴吧账号绑定到多个云签账号
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
- [x] 兼容部分原版插件
  - [x] [自动刷新贴吧列表](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ref)
  - [x] [名人堂](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_rank)
  - [x] [循环封禁](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_ban)
  - [x] ~~[删贴机](https://github.com/MoeNetwork/Tieba-Cloud-Sign/tree/master/plugins/ver4_review)（可能会拖很久甚至不会做）~~ 不再考虑跟进删贴机，有需要的请选择其他贴吧管理软件
  - [x] ?[知道商城抽奖](https://github.com/96dl/Tieba-Cloud-Sign-Plugins/blob/master/ver4_lottery/ver4_lottery_desc.php)
- [ ] 优化 PHP 原版的相关功能
- [x] 不再考虑 ~~通过读取 `config.php` 取得数据库连接信息~~
- [x] 自动清理解除绑定的账号的插件设置
- [x] ?打包/Docker/或者别的
- [x] ?自动化部署
- [ ] ?支持更多 Gorm 也支持的数据库
- [x] ?邮箱以外的推送方式
- [ ] 完善权限控制
- [x] 个人数据导出 (接口 `/passport/export` 已写好，但没想好如何处理好安全问题，当前所有接口都会自动删除 `bduss` 和 `stoken` 的值，但导出会不可避免地需要处理这个问题)，强烈建议阅读 readme.md 的 [备份](#备份) 部分
- [ ] ……更多的想起来再加

## 更多

更多的 todo 和已知 bug 请查看实时更新的 [Projects/TbSign➡️](https://github.com/users/BANKA2017/projects/4)

## 感谢

- [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)
- [Starry-OvO/aiotieba](https://github.com/Starry-OvO/aiotieba/)
