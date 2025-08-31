package rest

import "resty.dev/v3"

type REST struct {
	client *resty.Client
}

func NewREST() *REST {
	return &REST{
		client: resty.New(),
	}
}

func (r *REST) Close() error {
	return r.client.Close()
}

func (r *REST) Get(url string) (*resty.Response, error) {
	return r.client.R().Get(url)
}
