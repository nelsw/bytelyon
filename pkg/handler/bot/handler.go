package bot

import (
	"net/http"

	. "github.com/nelsw/bytelyon/pkg/api"
	"github.com/nelsw/bytelyon/pkg/db"
	. "github.com/nelsw/bytelyon/pkg/model"
)

// api/user/login
// api/user/reset
// api/user/signup
// api/user/token/:token

// api/bots/type/:type
// api/bots/type/:type/target/:target
// api/bots/type/:type/target/:target/id/:id

// Handler
func Handler(r Request) (Response, error) {

	r.Log()
	t, err := DetermineType(r.Query("type"))
	if err != nil {
		return r.BAD(err), nil
	}

	var tgt Target
	tgt, err = ConstructTarget(r.Query("target"))
	var bot = Bot[Type]{
		UserID: r.UserID(),
		T:      t,
		Target: tgt,
	}

	if err = bot.T.Validate(); err != nil {
		return r.BAD(err), nil
	}

	switch r.Method() {
	//case http.MethodDelete:
	//	return hanleDelete(r, b), nil
	//case http.MethodGet:
	//	return handleGet(r, b), nil
	case http.MethodPost:
		fallthrough
	case http.MethodPut:
		return handlePut(r), nil
	}

	return r.NI(), nil
}

func handleDelete(r Request, b Bot[Type]) Response {

	if err := db.Delete(b); err != nil {
		return r.BAD(err)
	}
	return r.NC()
}

func handleGet(r Request, b Bot[Type]) Response {
	arr, err := db.Query[Bot](b)
	if err != nil {
		return r.BAD(err)
	}
	return r.OK(arr)
}

func handlePut(r Request) Response {

	var b Bot[Bot[Type]]
	//if err := Body[Bot](r, b); err != nil {
	//	return r.BAD(err)
	//}

	b.UserID = r.UserID()
	b.BotID = NewULID()
	if err := db.Put(b); err != nil {
		return r.BAD(err)
	}
	return r.OK(b)
}
