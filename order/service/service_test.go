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

package service

import (
	_ "github.com/synthecypher/digota/product/service"
	_ "github.com/synthecypher/digota/sku/service"
)

import (
	"github.com/synthecypher/digota/config"
	"github.com/synthecypher/digota/locker"
	"github.com/synthecypher/digota/order/orderpb"
	"github.com/synthecypher/digota/payment/paymentpb"
	"github.com/synthecypher/digota/payment/service/providers"
	"github.com/synthecypher/digota/product"
	"github.com/synthecypher/digota/product/productpb"
	"github.com/synthecypher/digota/sku"
	"github.com/synthecypher/digota/sku/skupb"
	"github.com/synthecypher/digota/storage"
	"github.com/icrowley/fake"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"os"
	"testing"
	"time"
)

var db = "testing" + uuid.NewV4().String()

func TestMain(m *testing.M) {

	// setup
	if err := storage.New(config.Storage{
		Address:  []string{"localhost"},
		Handler:  "mongodb",
		Database: db,
	}); err != nil {
		panic(err)
	}

	// in-memory locker
	locker.New(config.Locker{})

	providers.New([]config.PaymentProvider{{Provider: "DigotaInternalTestOnly"}})

	retCode := m.Run()
	// teardown
	storage.Handler().DropDatabase(db)
	os.Exit(retCode)
}

func createDemoProduct() (*productpb.Product, error) {
	return product.Service().New(context.Background(), &productpb.NewRequest{
		Active:      true,
		Name:        fake.Sentences(),
		Description: fake.Sentences(),
		Shippable:   false,
		Url:         "http://" + fake.Characters() + ".com",
		Images:      []string{"http://" + fake.Characters() + ".com", "http://" + fake.Characters() + ".com"},
		Metadata: map[string]string{
			"key": "val",
		},
		Attributes: []string{"color"},
	})
}

func createSku(product *productpb.Product, currency paymentpb.Currency, active bool) (*skupb.Sku, error) {
	return sku.Service().New(context.Background(), &skupb.NewRequest{
		Parent:   product.Id,
		Name:     fake.Sentences(),
		Currency: currency,
		Active:   active,
		Price:    1500,
		Image:    "http://" + fake.Characters() + ".com",
		Inventory: &skupb.Inventory{
			Type:     skupb.Inventory_Finite,
			Quantity: 3,
		},
		Attributes: map[string]string{
			"color": "red",
		},
	})
}

func createOrder() (*orderpb.Order, error) {
	orderService := orderService{}

	// add product
	demoproduct, err := createDemoProduct()
	if err != nil {
		panic(err)
	}
	// add sku
	sku1, err := createSku(demoproduct, paymentpb.Currency_USD, true)
	if err != nil {
		panic(err)
	}
	// create new order
	// ok
	return orderService.New(context.Background(), &orderpb.NewRequest{
		Currency: paymentpb.Currency_USD,
		Items: []*orderpb.OrderItem{
			{
				Parent:   sku1.GetId(),
				Quantity: 2,
				Type:     orderpb.OrderItem_sku,
			},
			{
				Amount:      -1000,
				Description: "on the fly discount without parent",
				Currency:    paymentpb.Currency_USD,
				Type:        orderpb.OrderItem_discount,
			},
			{
				Amount:      50,
				Description: "Tax (Included)",
				Currency:    paymentpb.Currency_USD,
				Type:        orderpb.OrderItem_tax,
			},
		},
		Email: "yaron@digota.com",
		Shipping: &orderpb.Shipping{
			Name:  "Yaron Sumel",
			Phone: "+972 000 000 000",
			Address: &orderpb.Shipping_Address{
				Line1:      "Loren ipsum",
				City:       "San Jose",
				Country:    "USA",
				Line2:      "",
				PostalCode: "12345",
				State:      "CA",
			},
		},
	})
}

func TestOrders_GetNamespace(t *testing.T) {
	o := orders{}
	if o.GetNamespace() != "order" {
		t.FailNow()
	}
}

func TestOrder_GetNamespace(t *testing.T) {
	o := order{}
	if o.GetNamespace() != "order" {
		t.FailNow()
	}
}

func TestOrder_SetId(t *testing.T) {
	o := order{}
	o.SetId("1234-1234-1234-1234")
	if o.Id != "1234-1234-1234-1234" {
		t.FailNow()
	}
}

func TestOrder_SetCreated(t *testing.T) {
	o := order{}
	now := time.Now().Unix()
	o.SetCreated(now)
	if o.Created != now {
		t.FailNow()
	}
}

func TestOrder_SetUpdated(t *testing.T) {
	o := order{}
	now := time.Now().Unix()
	o.SetUpdated(now)
	if o.Updated != now {
		t.FailNow()
	}
}

func TestOrder_IsReturnable(t *testing.T) {
	o := order{}
	o.Amount = 1000
	o.Currency = paymentpb.Currency_USD
	o.Status = orderpb.Order_Created
	// should'nt be nil
	if err := o.IsReturnable(100); err == nil {
		t.Fatal(err)
	}
	o.Status = orderpb.Order_Paid
	// should'nt be nil
	if err := o.IsReturnable(100); err != nil {
		t.Fatal(err)
	}
	o.Amount = 100
	// should be nil
	if err := o.IsReturnable(100); err != nil {
		t.Fatal(err)
	}
}

func TestOrder_IsPayable(t *testing.T) {
	o := order{}
	o.Amount = 0
	o.Created = 0
	o.Status = orderpb.Order_Paid
	// should'nt be nil
	if err := o.IsPayable(); err == nil {
		t.Fatal(err)
	}
	o.Status = orderpb.Order_Created
	// should'nt be nil
	if err := o.IsPayable(); err == nil {
		t.Fatal(err)
	}
	o.Created = time.Now().Unix()
	// should'nt be nil
	if err := o.IsPayable(); err == nil {
		t.Fatal(err)
	}
	o.Amount = 100
	// should be nil
	if err := o.IsPayable(); err != nil {
		t.Fatal(err)
	}
}

func TestService_New(t *testing.T) {

	orderService := orderService{}

	// add product
	demoproduct, err := createDemoProduct()
	if err != nil {
		t.Fatal(err)
	}

	// add sku
	sku1, err := createSku(demoproduct, paymentpb.Currency_USD, true)
	if err != nil {
		t.Fatal(err)
	}

	// create new order
	// ok
	o, err := orderService.New(context.Background(), &orderpb.NewRequest{
		Currency: paymentpb.Currency_USD,
		Items: []*orderpb.OrderItem{
			{
				Parent:   sku1.GetId(),
				Quantity: 2,
				Type:     orderpb.OrderItem_sku,
			},
			{
				Amount:      -1000,
				Description: "on the fly discount without parent",
				Currency:    paymentpb.Currency_USD,
				Type:        orderpb.OrderItem_discount,
			},
			{
				Amount:      50,
				Description: "Tax (Included)",
				Currency:    paymentpb.Currency_USD,
				Type:        orderpb.OrderItem_tax,
			},
		},
		Email: "yaron@digota.com",
		Shipping: &orderpb.Shipping{
			Name:  "Yaron Sumel",
			Phone: "+972 000 000 000",
			Address: &orderpb.Shipping_Address{
				Line1:      "Loren ipsum",
				City:       "San Jose",
				Country:    "USA",
				Line2:      "",
				PostalCode: "12345",
				State:      "CA",
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	// check amount != amount*quantity-discount+tax
	if o.Amount != int64(sku1.Price)*2-1000+50 {
		t.Fatal()
	}

	// validation error
	if _, err := orderService.New(context.Background(), &orderpb.NewRequest{
		Currency: paymentpb.Currency_USD,
		Items: []*orderpb.OrderItem{
			{
				Parent:   "notuuid",
				Quantity: 2,
				Type:     orderpb.OrderItem_sku,
			},
		},
		Email: "yarondigota",
	}); err == nil {
		t.Fatal()
	}

	// not found
	if _, err := orderService.New(context.Background(), &orderpb.NewRequest{
		Currency: paymentpb.Currency_USD,
		Items: []*orderpb.OrderItem{
			{
				Parent:   uuid.NewV4().String(),
				Quantity: 2,
				Type:     orderpb.OrderItem_sku,
			},
		},
	}); err == nil {
		t.Fatal(err)
	}

	// not same currency
	if _, err := orderService.New(context.Background(), &orderpb.NewRequest{
		Currency: paymentpb.Currency_EUR,
		Items: []*orderpb.OrderItem{
			{
				Parent:   sku1.GetId(),
				Quantity: 2,
				Type:     orderpb.OrderItem_sku,
			},
		},
	}); err == nil {
		t.Fatal(err)
	}

}

func TestService_Get(t *testing.T) {

	orderService := orderService{}

	o, err := createOrder()

	if err != nil {
		t.Fatal(err)
	}

	o1, err := orderService.Get(context.Background(), &orderpb.GetRequest{
		Id: o.GetId(),
	})

	if err != nil {
		t.Fatal(err)
	}

	if o1.GetId() != o.GetId() {
		t.Fatal()
	}

	if _, err := orderService.Get(context.Background(), &orderpb.GetRequest{
		Id: "notvaliderr",
	}); err == nil {
		t.Fatal()
	}

	if _, err := orderService.Get(context.Background(), &orderpb.GetRequest{
		Id: uuid.NewV4().String(),
	}); err == nil {
		t.Fatal()
	}

}

func TestOrder_CalculateTotal(t *testing.T) {
	o := order{}
	o.Amount = 0
	o.Currency = paymentpb.Currency_USD

	orderItems0 := []*orderpb.OrderItem{
		{
			Amount:   100,
			Currency: paymentpb.Currency_AZN,
			Quantity: 2,
		},
	}

	if _, err := calculateTotal(o.Currency, orderItems0); err == nil {
		t.Fatal(err)
	}

	orderItems := []*orderpb.OrderItem{
		{
			Amount:   100,
			Currency: paymentpb.Currency_USD,
			Quantity: 2,
		},
		{
			Amount:   1000,
			Currency: paymentpb.Currency_USD,
			Quantity: 1,
		},
	}
	amount, err := calculateTotal(o.Currency, orderItems)
	if err != nil {
		t.Fatal(err)
	}

	if amount != 100*2+1000*1 {
		t.FailNow()
	}

}
