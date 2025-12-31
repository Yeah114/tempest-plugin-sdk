# tempest-plugin-sdk

这是 Tempest 的精简 Go 插件 SDK（给 `DynamicLoader` 用）。

## DynamicLoader 目录约定

运行时目录位于 `tempest_storage/lang/DynamicLoader/`：

- `code/`：插件源码（可选，便于你管理）
- `exe/`：插件编译后的可执行文件（DynamicLoader 从这里启动）

## 注意事项

- **不要向 stdout 打印日志**：`hashicorp/go-plugin` 默认会占用 stdout 做握手与协议通信；请把日志输出到 stderr（例如 `log.New(os.Stderr, ...)`）。

## 快速示例

SDK 内置了一个最小示例插件：`examples/echo`，会注册一个终端菜单命令 `echo`。

编译（示例，Linux）：

```bash
cd tempest-plugin-sdk/examples/echo
go build -o /root/tempest/tempest_storage/lang/DynamicLoader/exe/echo
```

配置（放到 `tempest_storage/config/...`，来源写 `DynamicLoader`）示例：

```json
{
  "名称": "echo",
  "描述": "dynamic echo example",
  "是否禁用": false,
  "版本": "0.0.1",
  "来源": "DynamicLoader",
  "配置": {
    "可执行文件": "echo"
  }
}
```

启动 Tempest 后，在终端输入 `echo hello` 应该能看到插件输出。

