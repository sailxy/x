# AWS S3 对象存储

`storage/s3` 提供 AWS S3 的对象上传和预签名上传地址生成能力。

迁移：将 `github.com/sailxy/x/aws/s3` 改为 `github.com/sailxy/x/storage/s3`。本包不提供跨云厂商统一对象存储接口。

```go
client := s3.New(s3.Config{Bucket: "example-bucket"})

object, err := client.PutObject(ctx, "uploads/a.txt", content)
upload, err := client.PresignPutObject(ctx, "uploads/a.txt")
```

AWS 凭据和区域沿用 AWS SDK 默认配置。
