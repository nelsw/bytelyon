package queue

import (
	"sync"

	"github.com/nelsw/bytelyon/apps/mux/internal/model"
)

type Queue struct {
	sync.Mutex
	ip map[int]bool
	ch chan *model.Bot
}

func New() *Queue {
	return &Queue{
		ip: make(map[int]bool),
		ch: make(chan *model.Bot),
	}
}

func (q *Queue) Send(bots ...*model.Bot) {
	for _, bot := range bots {
		q.ch <- bot
	}
}

func (q *Queue) Put(id int) bool {
	q.Lock()
	defer q.Unlock()
	if _, exists := q.ip[id]; exists {
		return false
	}
	q.ip[id] = true
	return true
}

func (q *Queue) Del(id int) {
	q.Lock()
	defer q.Unlock()
	delete(q.ip, id)
}

func (q *Queue) Close() {
	close(q.ch)
}

func (q *Queue) Chan() <-chan *model.Bot {
	return q.ch
}
