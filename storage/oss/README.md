# 阿里云 OSS 对象存储

`storage/oss` 提供阿里云 OSS 的签名 URL、前端 Post Policy 直传和对象下载能力。

迁移：将 `github.com/sailxy/x/aliyun/oss` 改为 `github.com/sailxy/x/storage/oss`。本包不提供跨云厂商统一对象存储接口。

```go
client, err := oss.New(oss.Config{
	Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
	BucketName:      "example-bucket",
	AccessKeyID:     "access-key-id",
	AccessKeySecret: "access-key-secret",
})

upload, err := client.SignURL("uploads/a.txt", oss.SignURLConfig{})
post, err := client.PostInfo("uploads/")
```
