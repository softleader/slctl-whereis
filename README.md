# slctl-whereis

The [slctl](https://github.com/softleader/slctl) plugin to find out where the SoftLeader member is

## Install

```sh
$ slctl plugin install github.com/softleader/slctl-whereis
```

## Usage

查看當日公司員工在哪兒

```sh
$ slctl whereis
```

可以使用員工姓名(模糊查詢)過濾資料

```sh
$ slctl whereis matt
```

傳入 `--from` 或 `--to` 可以用日期區間過濾, 日期格式為年月日, 支援格式可參考: [https://github.com/araddon/dateparse](https://github.com/araddon/dateparse), 同時也支援少數自然語言, 如: *today*, *yesterday*, *tomorrow*

```sh
$ slctl whereis -f yesterday
$ slctl whereis matt -f 20181201 -t 20181203
```

查詢結果預設顯示第一頁, 每頁顯示 20 筆資料, 可以傳入 `--page` 指定頁數或傳入 `--size` 指定一頁幾筆 (一頁筆數放很大則等於不分頁)

```sh
$ slctl whereis -s 1000
```

傳入 `--grep` 可以針對顯示的結果, 再做一次 regex 過濾, 類似 unix 系統的 `grep` 但可以跨系統使用

```sh
# 查詢 2019 年起到當日的資料, 但只顯示每個星期一的紀錄
$ slctl whereis matt -f 20190101 -t today --grep mon
```
