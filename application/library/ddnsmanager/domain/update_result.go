package domain

import "github.com/nging-plugins/ddnsmanager/application/library/ddnsmanager/domain/dnsdomain"

type UpdateResult struct {
	Provider string
	Updated  []*dnsdomain.Domain
	Error    error
}
