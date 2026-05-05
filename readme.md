# TbSign➡️

---

只是一个签到程序，可以配合[百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)使用，也可以独立运作

## ⚠ 注意

- 随缘维护，不保证不会封号，请**不要**用于主力账号！！！

## 启动参数

| flag                | env                  | default                  | description                                                                                                                     |
| :------------------ | :------------------- | :----------------------- | :------------------------------------------------------------------------------------------------------------------------------ |
| username            | tc_username          |                          | 数据库账号                                                                                                                      |
| pwd                 | tc_pwd               |                          | 数据库密码                                                                                                                      |
| ~~endpoint~~        | ~~tc_endpoint~~      | `127.0.0.1:3306`         | 数据库 `host:port` （已废弃）                                                                                                   |
| host                | tc_host              | `127.0.0.1:3306`         | 数据库 `host:port`                                                                                                              |
| db                  | tc_db                | `tbsign`                 | 数据库名称                                                                                                                      |
| db_tls              | tc_db_tls            | `false`                  | CA 证书的选项，请看 [CA 证书](#ca-证书) 部分                                                                                    |
| db_path             | tc_db_path           |                          | SQLite 文件目录                                                                                                                 |
| db_mode             | tc_db_mode           | `mysql`                  | 用于显式指定数据库类型，只用于 PostgreSQL (仅当值为 `pgsql` 或 `postgresql` 时有效，请看 [PostgreSQL](#postgresql) 部分)        |
| test                | tc_test              | `false`                  | 测试模式，此模式下不会运行计划任务                                                                                              |
| api                 | tc_api               | `false`                  | 是否启动 api                                                                                                                    |
| fe                  | tc_fe                | `false`                  | 是否启动自带前端，仅当 `api` 为 `true` 时可为 `true`                                                                            |
| address             | tc_address           | `:1323`                  | 后端运行地址                                                                                                                    |
| setup               |                      | `false`                  | 强制安装程序（可能会覆盖现有配置）                                                                                              |
| auto_install        | tc_auto_install      | `false`                  | 自动安装（仅当数据库不存在时安装）                                                                                              |
| admin_name          | tc_admin_name        |                          | 管理员账号，仅当 `auto_install` 为 `true` 时会用到                                                                              |
| admin_email         | tc_admin_email       |                          | 管理员邮箱，仅当 `auto_install` 为 `true` 时会用到                                                                              |
| admin_password      | tc_admin_password    |                          | 管理员密码，仅当 `auto_install` 为 `true` 时会用到                                                                              |
| no_proxy            |                      | `false`                  | 忽略环境变量中的代理配置                                                                                                        |
| allow_backup        | tc_allow_backup      | `false`                  | 允许用户批量导出/导入账号和贴吧列表，建议阅读 readme.md 的 [备份](#备份) 部分                                                   |
| data_encrypt_key    | tc_data_encrypt_key  |                          | 加密部分用户数据的密钥，使用 `base64url` 格式填写，建议阅读 readme.md 的 [加密](#加密) 部分                                     |
| data_encrypt_action |                      |                          | `encrypt` 或者 `decrypt`，用于加密/解密用户数据，处理完成后会自动退出，默认应当留空                                             |
| dns_addr            | tc_dns_addr          |                          | 手动设置 DNS 地址，特殊情况下使用，默认应当留空，建议阅读 readme.md 的 [网络](#网络) 部分                                       |
|                     | tc_docker_mode       |                          | 当值为 `true`(string) 且 `/.dockerenv` 存在时，激活 docker 模式，默认应当忽略，建议阅读 readme.md 的 [发布类型](#发布类型) 部分 |
| release_file_base   | tc_release_file_base | `share.ReleaseFilesPath` | 手动安装包的下载地址，特殊情况下使用，默认应当忽略，建议阅读 readme.md 的 [发布地址](#发布地址) 部分                            |
| release_api_base    | tc_release_api_base  | `share.ReleaseApiBase`   | 手动检查更新地址，特殊情况下使用，默认应当忽略，目前没有使用                                                                    |
| release_api_list    | tc_release_api_list  | `share.ReleaseApiList`   | 前端检查更新地址，特殊情况下使用，默认应当忽略                                                                                  |


- 不支持 `.env` 文件，请直接设置环境变量，使用顺序是 `flag` > `env` > `default`
- 有几个值不支持环境变量，必须手动操作
- 环境变量的 `false` 值可以是以下之一：`"0"`, `"false"`, `"off"`, `"disabled"`, `"no"`, `"n"`, `""`（空字符串），或不设置该环境变量
  - 以外的任意值都会被视为 `true`

示例

```shell
go run main.go -username=<dbUsername> -pwd=<DBPassword>
# or
./tbsign_go -username=<dbUsername> -pwd=<DBPassword>
# or https://github.com/cosmtrek/air
air -- -db_path=tbsign.db -test -api
```

## 网络

- 内嵌由 [curl](https://curl.se/docs/caextract.html) 提供的 Mozilla Root Store，当系统证书库不可用时会调用
  - 仓库自身不带 CA 库，请自行下载并放置在 `assets/ca/cacert.pem`，参考指令：`curl -o assets/ca/cacert.pem https://curl.se/ca/cacert.pem`
- 可以手动设置 DNS 服务器地址，支持的文本格式请参考 <https://pkg.go.dev/net#Dial>，示例：`8.8.8.8:53`

## 数据库

`db_mode`/`tc_db_mode` 的值为 `pgsql` 或 `postgresql` 时（不区分大小写）选择 PostgreSQL，否则：

当 `db_path`/`tc_db_path` 的值不为空字符串，就会使用 SQLite；否则使用 MySQL

MySQL 要求支持 [MySQL 窗口函数](https://dev.mysql.com/doc/refman/8.0/en/window-functions.html)

- MySQL >= [8.0](https://dev.mysql.com/doc/refman/8.0/en/window-functions.html)
- MariaDB >= [10.2](https://mariadb.com/kb/en/changes-improvements-in-mariadb-10-2/)
- TiDB >= [3.0](https://docs.pingcap.com/tidb/stable/release-3.0-ga)

### 超时

为了避免启动时超长时间的等待，设定了**初次**连接 `30s` 超时，超时后会报错退出

### *PostgreSQL

支持 PostgreSQL 是实验性的功能，不会被砍掉。下面是一些测试环境的细节

- 测试数据库使用 [Supabase](https://supabase.com/) 的云数据库
- 由于需要向前兼容，`db_mode` 无法用于另外两类数据库
- 开发者对 PostgreSQL 不了解，以前从未使用过 PostgreSQL

### CA 证书

适用于 `MySQL` 和 `PostgreSQL`

有的云服务，要求用户使用 TLS 连接到它们的数据库；有的用户对内网环境有特殊的安全需求，此时可能需要用到证书文件

#### MySQL TLS

对于 MySQL 可用于 `db_tls` 的值包括 `true`, `false`（默认值）, `skip-verify`, `preferred` 以及证书文件的路径，更多信息请参考 [go-sql-driver/mysql#tls](https://github.com/go-sql-driver/mysql?tab=readme-ov-file#tls)

例如下面的第二项是 Debian/Ubuntu 的目录；如果证书尚未被导入到系统，部署时就需要这样填写证书的实际目录

- 使用 docker 部署时请注意处理好文件目录问题
- 当 `db_tls=true` 时（不区分大小写），使用系统证书+Mozilla Root Store
- `db_tls` 类型为 `string`，cli 传参时不能像 `boolean` 类型那样忽略掉 `=true`

```shell
go run main.go -db_tls=true
# or...
go run main.go -db_tls=/etc/ssl/certs/ca-certificates.crt
```

#### PostgreSQL SSL Mode

对于 PostgreSQL 有 `disable`, `allow`, `prefer`（默认值）, `require`, `verify-ca`, `verify-full` 以及证书文件的路径，更多信息请参考 [LIBPQ-SSL-PROTECTION](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-PROTECTION)

用法跟 MySQL 差不多，证书使用默认的 `~/.postgres/root.crt` 或者自行准备的文件

- 使用 docker 部署时请注意处理好文件目录问题

## 安装

- 迁移法 (仅 MySQL)
  - 直接安装 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)
  - 使用原数据库配置启动
- 全新安装 (MySQL, SQLite, PostgreSQL)
  - 手动安装
    - 启动程序，添加 `flag` `-setup` (示例: `./tbsign_go -db_path=tbsign.db -api -address=:8080 -setup`)
    - 根据文字提示完成自动安装流程（不同情况下文字提示可能会略有不同）

      ```plaintext
      ➜  tbsign_go git:(master) ✗ ./tbsign_go -db_path=tbsign1.db -api -setup
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
    - 启动程序，添加 `flags` 或 环境变量 (示例: `./tbsign_go -db=tbsign -username tcdb -pwd tcdb_password -api -address=:8080 -auto_install -admin_name=a -admin_email=a@a.a -admin_password=a`)

      - 担心 log 泄露信息的此时可以设随机值，等到安装完成后再登录修改
      - 除非数据库被删除，否则 用户名/邮箱/密码 仅在首次开启时会用到
    - \*(选做) 登录后台修改信息，开启自带插件

\* 注：`setup` 和 `auto_install` 不可同时为 `true`

### 兼容性

- [⚠️] 通过 *迁移法* 启动的程序能够与 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/) 共存，但部分插件没有对应的 PHP 版本
- [❌] 通过 *全新安装* 启动的程序无法兼容 [百度贴吧云签到](https://github.com/MoeNetwork/Tieba-Cloud-Sign/)，因为缺少部分数据表和设置选项
- [❌] MySQL 不支持无法使用 [MySQL 窗口函数](https://dev.mysql.com/doc/refman/8.0/en/window-functions.html) 的发行版，使用了**不支持**的函数 `ROW_NUMBER()`
- [⚠️] 开发环境使用的 MySQL 版本号为 `8.0` 或更高，其他 MySQL 发行版请自行测试是否可用
- [⚠️] PostgreSQL 不保证可用
- [⚠️] Go 版本为 `go1.25.6`，请自行检查[兼容性](https://tip.golang.org/wiki/MinimumRequirements)

## 前端

➡️ <https://github.com/BANKA2017/tbsign_go_fe>

### 嵌入式前端

此类部署需要用到其他仓库，并且需要手动或自动添加文件，如果不知道是怎么回事，请忽略本节，使用前后分离部署，或直接使用 build.sh 打包

- 使用嵌入式前端前，请设置环境变量

  ```env
  NUXT_BASE_PATH="/api"
  NUXT_USE_COOKIE_TOKEN="1"
  ```

- 在前端仓库的目录执行 `yarn run generate` 后将 `.output/public` 目录的文件拷贝到 `/assets/dist` 内
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
  - 对应插件的页面，插件名使用 `PluginInfo.PluginNameFE`
    - 在目录 `app/pages/plugin/` 添加对应插件的页面 vue 文件，文件名必须使用 `插件名.vue`
    - 或者在目录 `app/pages/plugin/` 添加对应 `插件名` 的目录，目录里面的 `index.vue` 会被自动导入
  - *如果对变量设置表单有特殊要求，还需要修改 `admin/system.vue` 的内容
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

### *Pure Go

这是一项试验中的功能，可能会引发未知的问题

SQLite 驱动使用 [gitlab.com:cznic/sqlite](https://gitlab.com/cznic/sqlite) (modernc.org/sqlite)，Gorm 包装使用 [glebarez/sqlite](https://github.com/glebarez/sqlite)

[受支持的版本](https://pkg.go.dev/modernc.org/sqlite#hdr-Supported_platforms_and_architectures)在编译时禁用 CGO 即可

### CGO

> go-sqlite3 is cgo package. If you want to build your app using go-sqlite3, you need gcc. However, after you have built and installed go-sqlite3 with `go install github.com/mattn/go-sqlite3` (which requires gcc), you can build your app without relying on gcc in future.

#### musl-gcc

如有静态编译的需求（如下文的 Docker 使用 Alpine 默认不使用 glibc），可以使用 `musl-gcc`，参考命令

```bash
# apt install -y musl-dev

CGO_ENABLED=1 CC=musl-gcc go build -ldflags "-linkmode external -extldflags -static" -tags netgo

# 或者 docker 一步到位
docker run --rm -v $(pwd):/app/tbsign -e EXTERNAL_LDFLAGS="-linkmode external -extldflags -static" -w /app/tbsign alpine:3 sh -c "apk add --no-cache --upgrade nodejs npm go git gcc musl-dev && npm install -g corepack && corepack enable && chmod +x build.sh && sh ./build.sh"

# root@rpi4:~# ldd tbsign_go
# not a dynamic executable
```

#### zig

用 [zig](https://ziglang.org/download/) 也不错

```bash
# https://ziglang.org/download/
# zig 交叉编译也是可以的
CC="zig cc -target x86_64-linux-musl" GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -tags netgo
CC="zig cc -target x86_64-windows-gnu -O2" GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -tags netgo -ldflags "-linkmode external"
```

#### xgo

或者使用 [xgo](https://github.com/techknowlogick/xgo)

```bash
# go install src.techknowlogick.com/xgo@v1.8.1-0.20250401170454-4b368d8a5afa
# docker pull ghcr.io/techknowlogick/xgo:go-1.25.6
CGO_ENABLED=1 $HOME/go/bin/xgo -go go-1.25.6 -tags netgo --targets=windows/amd64,darwin/amd64,darwin/arm64 ./
```

#### glibc

任何时候都不建议静态编译 glibc，应该找个带旧版本 glibc 的系统来编译，可用版本参考 [pypa/manylinux](https://github.com/pypa/manylinux)，不建议做交叉编译

```bash
# 这里做静态编译
GOOS=linux GOARCH=arm64 CGO_LDFLAGS="-static" CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static" -tags "netgo sqlite_omit_load_extension osusergo"
```

### build.sh

简单写了个编译脚本，存放在 `~/build.sh`

### Docker

➡️ [ghcr.io/banka2017/tbsign_go](https://github.com/BANKA2017/tbsign_go/pkgs/container/tbsign_go)

- Dockerfile 在 `~/docker`，官方发布的镜像只会支持 `linux/arm64` 和 `linux/amd64`
- docker-compose.yml 文件位于 `~/docker/docker-compose.yml`

Docker 版不支持在设置页面下载更新

参考启动命令

```bash
docker run -d --restart unless-stopped -v ./db:/app/tbsign/db -p 8080:1323 ghcr.io/banka2017/tbsign_go:master
```

- 默认数据库目录在 `/app/tbsign/db/tbsign_go.db`，
  - 如果使用 `SQLite` 建议将 `/app/tbsign/db` 映射到宿主机
  - 如果使用 `MySQL` 或 `PostgreSQL` 请手动将环境变量 `tc_db_path` 的值设为空字符串，覆盖原镜像的默认值
- 开放端口号为 `1323`

### 版本号

格式为 `tbsign_go.<YYYYMMDD>.<BACKEND_COMMIT_HASH[0:7]>.<FRONTEND_COMMIT_HASH[0:7]>.<OS>-<ARCH>` ，如果是用于 Windows 系统的可执行文件还会有 `.exe` 拓展名

官方提供的可执行文件包含 linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64 五个版本，手动触发编译的 Actions 任务

~~\* 由于添加 `musl-libc` 静态编译产物比较麻烦，所以不会发布 `musl` 版的可执行文件，如有需要可以解包 Docker 镜像~~

仓库发布的 Linux 版自带 `musl-libc`，无需处理 `libc` 问题

其它系统请自行编译运行

## 更新

### 发布类型

由 `share.BuildPublishType` 的值决定，具体情况如下表

| 类型         | 发布方式                                                                             | 支持更新自身文件 |
| :----------- | :----------------------------------------------------------------------------------- | :--------------- |
| `source`     | 通过源码直接编译的默认值                                                             | ❌                |
| `binary`     | 通过 [GitHub Releases](https://github.com/BANKA2017/tbsign_go/releases) 发布         | ✅                |
| `docker`     | 通过 [ghcr.io](https://github.com/BANKA2017/tbsign_go/pkgs/container/tbsign_go) 发布 | ❌                |
| 其他自定义值 | 由第三方打包者自行决定                                                               | ⚠️ 由打包者决定   |

### 发布地址

由 `share.ReleaseFilesPath` 决定，第三方打包者可以自行修改，目录的格式参考 GitHub Releases

## 备份

从 [tbsign_go.20241203.f7b5434.881b23b](https://github.com/BANKA2017/tbsign_go/tree/tbsign_go.20241203.f7b5434.881b23b) 版本起，支持用户导出/导入部分账号数据。

此前，tbsign_go 设计上不允许 API 返回 `bduss` 和 `stoken`，只能通过查看数据库获取，确保即使站点账号被盗，仍无法从 tbsign_go 导出敏感数据。此版本后，若站点开放备份功能，用户登录后可以导出这些数据。

为平衡**安全**与**便利**，系统设置页面（网页）新增了备份功能开关，并添加了启动时确定是否启用的总开关（cli）。未启用的站点将禁用相关 API，是否开启功能由站点管理员自行决定。

[加密](#加密)数据会在解密后以**明文**的形式被导出，需要注意传输和存储的安全

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

### 设置

用户的个人设置 (表 `tc_users_options` ) 以及在各个插件的设置都可以导出

### 插件备份

插件数据的导出由各个插件自行实现，部分插件可能不支持数据导出

## *加密

支持 数据加密 是实验性的功能，不会被砍掉，但前期可能会有严重的 bug 和漏洞，仅供尝鲜

为了保护用户数据，支持使用 `aes-256-gcm` 加密部分用户数据，目前已经支持

- [x] BDUSS
- [x] Stoken
- [x] Bark key
- [x] Ntfy topic
- [x] PushDeer key

加密后的数据格式前 12 位为 `iv`，其余部分为密文

请在加解密数据前备份数据库，以免发生意外

未来可能会支持加密更多的数据，升级前需要先将数据解密，避免出错，升级完成后重新加密

## *每日签到报告

每日报告 是实验性的功能，可能无法正常使用

在指定时间以后检查各个账号的签到情况，并且发送到开启报告的用户的默认推送渠道

目前支持推送的信息包括：

- [x] 各个贴吧账号的 成功/失败/待签/忽略 数量
- [x] 各个贴吧账号的 BDUSS 有效性

目前有一些值是硬编码到软件的，包括：

- 单次检查用户数：`100`，每分钟检查 100 个站点账号，一个站点账号可能会有多个贴吧账号，此时会一次性导出，所以会有卡顿崩溃的可能性
- ~~账号 BDUSS 无效阈值：`0.95`，由于错误代码 `1` 存在误报的可能，所以需要出错贴吧数量达到总贴吧数量的 `95%` 才会判断为 BDUSS 失效。但在极端情景（如只关注了一个贴吧，这个贴吧签到失败得到错误代码 `1`，此时失败率 `100%`）仍存在误报的可能性~~
  - 目前使用当天的签到记录中有无 `110000` 来判断是否登出

## 重签

重签使用 [截断二进制指数退避](https://en.wikipedia.org/wiki/Exponential_backoff)，默认最大退避间隔为 1 分钟（跟原版一样）

可以通过修改 `go_recheck_in_max_interval`（系统管理 -> 签到 -> 最大重签间隔 (分钟)）调整间隔，最小间隔为 1 分钟，如果间隔设置得过大可能会导致极端情况下无法完成全部重签任务

## 已知问题

- [x] 不支持限制单次签到贴吧总数，会一次性全部签完
- [x] 没有灵活的请求间隔和请求头模拟，有封号的风险
- [x] 只有签到这一个功能，对重签的处理约等于没有
- [x] 混乱的输出，随处可见的打点
- [x] 没有进行任何优化
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
