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

package main

import (
	"github.com/synthecypher/digota/order/orderpb"
	"github.com/synthecypher/digota/sdk"
	"golang.org/x/net/context"
	"log"
)

func main() {

	c, err := sdk.NewClient("localhost:3051", &sdk.ClientOpt{
		InsecureSkipVerify: false,
		ServerName:         "server.com",
		CaCrt:              "out/ca.crt",
		Crt:                "out/client.com.crt",
		Key:                "out/client.com.key",
	})

	if err != nil {
		panic(err)
	}

	defer c.Close()

	// Pay order
	l, err := orderpb.NewOrderServiceClient(c).List(context.Background(), &orderpb.ListRequest{
		Limit: 4,
		Page:  0,
	})

	if err != nil {
		panic(err)
	}

	for _, order := range l.Orders {
		log.Println(order.GetId(), order.GetCreated())
	}

}
