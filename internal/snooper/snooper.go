package snooper

import (
	"time"

	"github.com/nelsw/bytelyon/pkg/shopify"
	"github.com/nelsw/bytelyon/pkg/store"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	tkn, store string
	stop, done bool
}

func New(tkn, store string) *Worker {
	return &Worker{tkn, store, false, true}
}

func (w *Worker) Start() {
	log.Info().Msg("starting")
	for !w.stop {
		w.work()
		w.sleep()
	}
}

func (w *Worker) Stop() {
	w.stop = true
}

func (w *Worker) work() {
	log.Info().Msg("working")

	var err error
	var orderDB *store.DB[shopify.Order]
	var customerDB *store.DB[shopify.Customer]

	if customerDB, err = store.New[shopify.Customer]("customers.json"); err != nil {
		panic(err)
	} else if orderDB, err = store.New[shopify.Order]("orders.json"); err != nil {
		panic(err)
	}
	defer func() {
		customerDB.Close()
		orderDB.Close()
	}()

	var orders shopify.Orders
	if orders, err = shopify.GetOrders(w.tkn, w.store, time.Now().Add(-24*time.Hour), time.Now()); err != nil {
		panic(err)
	}
	for _, order := range orders {
		orderDB.Put(order.ID, order)
		customerDB.Put(order.Customer.ID, order.Customer)
	}
}

func (w *Worker) sleep() {
	log.Info().Msg("sleeping")
	for i := 0; i < 60 && !w.stop; i++ {
		time.Sleep(time.Hour * 24)
	}
}
