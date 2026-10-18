package oss

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	aliyunoss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// BucketState uses native V2 operations; no bucket configuration is changed.
func (c *Client) BucketState(ctx context.Context) (*BucketState, error) {
	acl, err := c.sdkClient.GetBucketAcl(ctx, &aliyunoss.GetBucketAclRequest{Bucket: aliyunoss.Ptr(c.bucketName)}, expectXMLRoot("AccessControlPolicy"))
	if err != nil {
		return nil, fmt.Errorf("get bucket ACL: %w", err)
	}
	version, err := c.sdkClient.GetBucketVersioning(ctx, &aliyunoss.GetBucketVersioningRequest{Bucket: aliyunoss.Ptr(c.bucketName)}, expectXMLRoot("VersioningConfiguration"))
	if err != nil {
		return nil, fmt.Errorf("get bucket versioning: %w", err)
	}
	return &BucketState{ACL: aliyunoss.ToString(acl.ACL), VersioningStatus: aliyunoss.ToString(version.VersionStatus)}, nil
}
func expectXMLRoot(root string) func(*aliyunoss.Options) {
	return func(o *aliyunoss.Options) {
		o.ResponseHandlers = append(o.ResponseHandlers, func(response *http.Response) error {
			if response.StatusCode < 200 || response.StatusCode >= 300 {
				return nil
			}
			// V2 accepts unknown roots as empty results; do not confuse one with unversioned.
			body, err := io.ReadAll(io.LimitReader(response.Body, 64*1024+1))
			closeErr := response.Body.Close()
			response.Body = io.NopCloser(bytes.NewReader(body))
			if err != nil {
				return fmt.Errorf("read bucket response: %w", err)
			}
			if closeErr != nil {
				return fmt.Errorf("close bucket response: %w", closeErr)
			}
			if len(body) > 64*1024 {
				return fmt.Errorf("bucket response exceeds size limit")
			}
			var shape struct{ XMLName xml.Name }
			if err := xml.Unmarshal(body, &shape); err != nil {
				return fmt.Errorf("validate bucket response: %w", err)
			}
			if shape.XMLName.Local != root {
				return fmt.Errorf("unexpected bucket XML root")
			}
			return nil
		})
	}
}
