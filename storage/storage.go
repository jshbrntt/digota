//     Digota <http://digota.com> - eCommerce microservice
//     Copyright (C) 2017  Yaron Sumel <yaron@digota.com>. All Rights Reserved.
//
//     This program is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published
//     by the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     This program is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.
//
//     You should have received a copy of the GNU Affero General Public License
//     along with this program.  If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"errors"
	"github.com/synthecypher/digota/config"
	"github.com/synthecypher/digota/storage/handlers/mongo"
	"github.com/synthecypher/digota/storage/object"
)

const (
	mongodbHandler handlerName = "mongodb"
)

type (
	handlerName string
	// Interface defines the base functionality which any storage handler
	// should implement to become valid storage handler
	Interface interface {
		Prepare() error
		Close() error
		DropCollection(db string, doc object.Interface) error
		DropDatabase(db string) error
		One(doc object.Interface) error
		List(docs object.Interfaces, opt object.ListOpt) (int, error)
		ListParent(parent string, docs object.Interfaces) error
		Insert(doc object.Interface) error
		Update(doc object.Interface) error
		Remove(doc object.Interface) error
	}
)

var handler Interface

// New creates storage handler from config.Storage and prepare it for use
// returns error if something went wrong during the preparations
func New(storageConfig config.Storage) error {
	// create handler based on the storage config
	switch handlerName(storageConfig.Handler) {
	case mongodbHandler:
		handler = mongo.NewHandler(storageConfig)
	default:
		return errors.New("Invalid storage handler `" + storageConfig.Handler + "`")
	}
	// prepare handler
	return handler.Prepare()
}

// Handler returns the registered storage handler
func Handler() Interface {
	return handler
}
