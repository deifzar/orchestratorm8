package db8

import (
	"deifzar/orchestratorm8/pkg/model8"

	"github.com/gofrs/uuid/v5"
)

type Db8Domain8Interface interface {
	InsertDomain(model8.PostDomain8) (bool, error)
	GetAllDomain() ([]model8.Domain8, error)
	GetOneDomain(uuid.UUID) (model8.Domain8, error)
	GetAllEnabled() ([]model8.Domain8, error)
	ExistEnabled() bool
	UpdateDomain(uuid.UUID, model8.PostDomain8) (model8.Domain8, error)
	DeleteDomain(uuid.UUID) (bool, error)
}
