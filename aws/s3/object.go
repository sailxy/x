package s3

type Object struct {
	Key  string
	ETag string
}

type PresignRequest struct {
	URL string
}
