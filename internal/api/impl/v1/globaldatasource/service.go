// Copyright 2021 The Perses Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package globaldatasource

import (
	"fmt"

	"github.com/perses/common/etcd"
	"github.com/perses/perses/internal/api/interface/v1/globaldatasource"
	"github.com/perses/perses/internal/api/shared"
	"github.com/perses/perses/internal/api/shared/schemas"
	"github.com/perses/perses/internal/api/shared/validate"
	"github.com/perses/perses/pkg/model/api"
	v1 "github.com/perses/perses/pkg/model/api/v1"
	"github.com/sirupsen/logrus"
)

type service struct {
	globaldatasource.Service
	dao globaldatasource.DAO
	sch schemas.Schemas
}

func NewService(dao globaldatasource.DAO, sch schemas.Schemas) globaldatasource.Service {
	return &service{
		dao: dao,
		sch: sch,
	}
}

func (s *service) Create(entity api.Entity) (interface{}, error) {
	if datasourceObject, ok := entity.(*v1.GlobalDatasource); ok {
		return s.create(datasourceObject)
	}
	return nil, fmt.Errorf("%w: wrong entity format, attempting GlobalDatasource format, received '%T'", shared.BadRequestError, entity)
}

func (s *service) create(entity *v1.GlobalDatasource) (*v1.GlobalDatasource, error) {
	if err := s.validate(entity); err != nil {
		return nil, fmt.Errorf("%w: %s", shared.BadRequestError, err)
	}
	// Update the time contains in the entity
	entity.Metadata.CreateNow()
	if err := s.dao.Create(entity); err != nil {
		if etcd.IsKeyConflict(err) {
			logrus.Debugf("unable to create the GlobalDatasource %q. It already exits", entity.Metadata.Name)
			return nil, shared.ConflictError
		}
		logrus.WithError(err).Errorf("unable to perform the creation of the GlobalDatasource %q, something wrong with etcd", entity.Metadata.Name)
		return nil, shared.InternalError
	}
	return entity, nil
}

func (s *service) Update(entity api.Entity, parameters shared.Parameters) (interface{}, error) {
	if DatasourceObject, ok := entity.(*v1.GlobalDatasource); ok {
		return s.update(DatasourceObject, parameters)
	}
	return nil, fmt.Errorf("%w: wrong entity format, attempting GlobalDatasource format, received '%T'", shared.BadRequestError, entity)
}

func (s *service) update(entity *v1.GlobalDatasource, parameters shared.Parameters) (*v1.GlobalDatasource, error) {
	if err := s.validate(entity); err != nil {
		return nil, fmt.Errorf("%w: %s", shared.BadRequestError, err)
	}
	if entity.Metadata.Name != parameters.Name {
		logrus.Debugf("name in Datasource %q and name from the http request %q don't match", entity.Metadata.Name, parameters.Name)
		return nil, fmt.Errorf("%w: metadata.name and the name in the http path request don't match", shared.BadRequestError)
	}
	// find the previous version of the Datasource
	oldEntity, err := s.Get(parameters)
	if err != nil {
		return nil, err
	}
	oldObject := oldEntity.(*v1.GlobalDatasource)
	entity.Metadata.Update(oldObject.Metadata)
	if err := s.dao.Update(entity); err != nil {
		logrus.WithError(err).Errorf("unable to perform the update of the GlobalDatasource %q, something wrong with etcd", entity.Metadata.Name)
		return nil, shared.InternalError
	}
	return entity, nil
}

func (s *service) Delete(parameters shared.Parameters) error {
	if err := s.dao.Delete(parameters.Name); err != nil {
		if etcd.IsKeyNotFound(err) {
			logrus.Debugf("unable to find the Datasource %q", parameters.Name)
			return shared.NotFoundError
		}
		logrus.WithError(err).Errorf("unable to delete the GlobalDatasource %q, something wrong with etcd", parameters.Name)
		return shared.InternalError
	}
	return nil
}

func (s *service) Get(parameters shared.Parameters) (interface{}, error) {
	entity, err := s.dao.Get(parameters.Name)
	if err != nil {
		if etcd.IsKeyNotFound(err) {
			logrus.Debugf("unable to find the Datasource %q", parameters.Name)
			return nil, shared.NotFoundError
		}
		logrus.WithError(err).Errorf("unable to find the previous version of the GlobalDatasource %q, something wrong with etcd", parameters.Name)
		return nil, shared.InternalError
	}
	return entity, nil
}

func (s *service) List(q etcd.Query, _ shared.Parameters) (interface{}, error) {
	dtsList, err := s.dao.List(q)
	if err != nil {
		return nil, err
	}
	dtsQuery := q.(*globaldatasource.Query)
	return v1.FilterDatasource(dtsQuery.Kind, dtsQuery.Default, dtsList), nil
}

func (s *service) validate(entity *v1.GlobalDatasource) error {
	var list []*v1.GlobalDatasource
	if entity.Spec.Default {
		var err error
		list, err = s.dao.List(&globaldatasource.Query{})
		if err != nil {
			logrus.WithError(err).Errorf("unable to get the list of the global datasource")
			return err
		}
	}
	return validate.Datasource(entity, list, s.sch)
}
