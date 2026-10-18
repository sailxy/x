# 阿里云 OSS 对象存储

`storage/oss` 提供阿里云 OSS 的签名 URL、前端 Post Policy 直传和对象下载能力。

迁移：将 `github.com/sailxy/x/aliyun/oss` 改为 `github.com/sailxy/x/storage/oss`。本包不提供跨云厂商统一对象存储接口。

```go
client, err := oss.New(oss.Config{
	Region:          "cn-hangzhou",
	Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
	BucketName:      "example-bucket",
	AccessKeyID:     "access-key-id",
	AccessKeySecret: "access-key-secret",
})

upload, err := client.SignURL("uploads/a.txt", oss.SignURLConfig{})
post, err := client.SignPostPolicy("uploads/", oss.PostPolicyConfig{
	ExpiredInSec:    600,
	MaxFileSize:     20 * 1024 * 1024,
	ForbidOverwrite: true,
	ObjectACL:       "private",
})
```

浏览器 POST 授权统一使用 V4 `SignPostPolicy`，必须配置 Region，并显式指定有效期和最大文件尺寸。原 `PostInfo` 已移除，旧调用必须升级；前端使用返回的 URL、Fields，并添加文件 key 和 file 表单项。
