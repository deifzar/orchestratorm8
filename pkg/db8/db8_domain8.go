package db8

import (
	"database/sql"
	"deifzar/orchestratorm8/pkg/log8"
	"deifzar/orchestratorm8/pkg/model8"

	"github.com/gofrs/uuid/v5"
	_ "github.com/lib/pq"
)

type Db8Domain8 struct {
	Db *sql.DB
}

func NewDb8Domain8(db *sql.DB) Db8Domain8Interface {
	return &Db8Domain8{Db: db}
}

func (m *Db8Domain8) DeleteDomain(id uuid.UUID) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	_, err = tx.Exec("DELETE FROM cptm8domain WHERE id = $1", id)
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	return true, nil
}

func (m *Db8Domain8) GetAllDomain() ([]model8.Domain8, error) {
	query, err := m.Db.Query("SELECT id, name, companyname, enabled FROM cptm8domain")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return []model8.Domain8{}, err
	}
	var domains []model8.Domain8
	if query != nil {
		for query.Next() {
			var (
				id          uuid.UUID
				name        string
				companyname string
				enabled     bool
			)
			err := query.Scan(&id, &name, &companyname, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return nil, err
			}
			d := model8.Domain8{Id: id, Name: name, Companyname: companyname, Enabled: enabled}
			domains = append(domains, d)
		}
	}
	return domains, nil
}

func (m *Db8Domain8) GetAllEnabled() ([]model8.Domain8, error) {
	query, err := m.Db.Query("SELECT id, name, companyname, enabled FROM cptm8domain WHERE enabled = true")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return []model8.Domain8{}, err
	}
	var domains []model8.Domain8
	if query != nil {
		for query.Next() {
			var (
				id          uuid.UUID
				name        string
				companyname string
				enabled     bool
			)
			err := query.Scan(&id, &name, &companyname, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return nil, err
			}
			d := model8.Domain8{Id: id, Name: name, Companyname: companyname, Enabled: enabled}
			domains = append(domains, d)
		}
	}
	return domains, nil
}

func (m *Db8Domain8) ExistEnabled() bool {
	err := m.Db.QueryRow("SELECT id, name, companyname, enabled FROM cptm8domain WHERE enabled = true").Scan()
	if err != nil {
		if err != sql.ErrNoRows {
			log8.BaseLogger.Debug().Stack().Msg(err.Error())
		}
		return false
	}
	return true
}

func (m *Db8Domain8) GetOneDomain(id uuid.UUID) (model8.Domain8, error) {
	query, err := m.Db.Query("SELECT id, name, companyname, enabled FROM cptm8domain WHERE id = $1", id)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Domain8{}, err
	}
	var domain model8.Domain8
	if query != nil {
		for query.Next() {
			var (
				id          uuid.UUID
				name        string
				companyname string
				enabled     bool
			)
			err := query.Scan(&id, &name, &companyname, &enabled)
			if err != nil {
				log8.BaseLogger.Debug().Msg(err.Error())
				return model8.Domain8{}, err
			}
			domain = model8.Domain8{Id: id, Name: name, Companyname: companyname, Enabled: enabled}
		}
	}
	return domain, nil
}

func (m *Db8Domain8) InsertDomain(post model8.PostDomain8) (bool, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	stmt, err := tx.Prepare("INSERT INTO cptm8domain(name, companyname, enabled) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(post.Name, post.Companyname, post.Enabled)
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return false, err2
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return false, err2
	}
	return true, nil
}

func (m *Db8Domain8) UpdateDomain(id uuid.UUID, post model8.PostDomain8) (model8.Domain8, error) {
	tx, err := m.Db.Begin()
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Domain8{}, err
	}

	stmt, err := tx.Prepare("UPDATE cptm8domain SET name = $1, companyname = $2, enabled = $3 WHERE id = $4")
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Domain8{}, err
	}
	_, err2 := stmt.Exec(post.Name, post.Companyname, post.Enabled, id)
	if err2 != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err2.Error())
		return model8.Domain8{}, err2
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Domain8{}, err
	}
	var d model8.Domain8
	d, err = m.GetOneDomain(id)
	if err != nil {
		log8.BaseLogger.Debug().Msg(err.Error())
		return model8.Domain8{}, err
	}
	return d, nil
}
