package model

import (
	"testing"
)

func Test_User(t *testing.T) {

	//db.Create(&User{})
	//
	//exp := NewUser()
	//err := db.Put(exp)
	//assert.NoError(t, err)
	//
	//var act User
	//act, err = db.Get[User](&User{ID: exp.ID})
	//assert.NoError(t, err)
	//assert.Equal(t, exp.ID, act.ID)
	//
	//err = db.Delete(&User{ID: exp.ID})
	//assert.NoError(t, err)
	//
	//act, err = db.Get[User](&User{ID: exp.ID})
	//assert.Empty(t, act)

	//for i := 0; i < 50; i++ {
	//	db.Put(&User{ID: NewULID()})
	//}

	//out, err := db.Scan[User](&User{})
	//assert.NoError(t, err)
	////assert.NotEmpty(t, out)
	////util.PrettyPrintln(out)
	//fmt.Println(len(out))
	//for i := 0; i < 100; i++ {
	//	u := out[i]
	//	bots, berr := db.Query[Bot](Bot{Type: SearchBotType, UserID: u.ID})
	//	if berr != nil {
	//		t.Errorf("failed to query bots for user %s: %v", u.ID, berr)
	//		continue
	//	}
	//	util.PrettyPrintln(bots)
	//}
}
