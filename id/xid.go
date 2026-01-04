package id

import "github.com/rs/xid"

// NewXID returns a new XID.
// xid is a globally unique id generator thought for the web.
func NewXID() (string, error) {
	return xid.New().String(), nil
}
