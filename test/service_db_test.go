package test

//
//func Test_DB_Find(t *testing.T) {
//
//	user := MakeUser()
//	assert.NoError(t, db.Put(user))
//
//	address := gofakeit.Email()
//	assert.NoError(t, db.Put(Email{user, address}))
//
//	email, err := db.Get[Email](Email{Address: address})
//	assert.NoError(t, err)
//	assert.Equal(t, address, email.Address)
//	assert.Equal(t, user.ID, email.User.ID)
//}
//
//func Test_DB_Query(t *testing.T) {
//
//	users, scanErr := db.Scan[User](User{})
//	assert.NoError(t, scanErr)
//	assert.NotEmpty(t, users)
//
//	user := users[0]
//
//	bots, queryErr := db.Query[Bot](Bot{User: User{ID: user.ID}})
//	assert.NoError(t, queryErr)
//	assert.NotEmpty(t, bots)
//	util.PrettyPrintln(bots)
//
//	for _, b := range bots {
//		assert.Equal(t, user.ID, b.User.ID)
//	}
//}
//
//func Test_DB_Scan(t *testing.T) {
//	arr, err := db.Scan[Bot](Bot{})
//	assert.NoError(t, err)
//	assert.NotEmpty(t, arr)
//	util.PrettyPrintln(arr)
//}
//
//func Test_DB_Save(t *testing.T) {
//
//	var users = []User{
//		MakeUser(),
//		MakeUser(),
//	}
//
//	assert.NoError(t, db.Put(users[0]))
//	assert.NoError(t, db.Put(users[1]))
//
//	for i := 0; i < 10; i++ {
//
//		bot := Bot{
//			User:      users[i%2],
//			BlackList: []string{fake.DomainName()},
//			Frequency: time.Hour * time.Duration(fake.Uint8()),
//			Target:    fake.URL(),
//		}
//
//		if i < 3 {
//			bot.Type = SearchBotType
//			bot.Headless = i%2 == 0
//			db.Put(bot)
//		} else if i < 5 {
//			bot.Type = SitemapBotType
//			db.Put(bot)
//		} else {
//			bot.Type = NewsBotType
//			db.Put(bot)
//		}
//	}
//
//	arr, err := db.Scan[Bot](Bot{Type: NewsBotType})
//	assert.NoError(t, err)
//	assert.NotEmpty(t, arr)
//	util.PrettyPrintln(arr)
//}
//
//func Test_DB_Wipe(t *testing.T) {
//
//	user := User{ID: uuid.New().String()}
//	db.Put(&user)
//
//	arr, err := db.Get[User](user)
//	assert.NoError(t, err)
//	assert.NotEmpty(t, arr)
//
//	err = db.Delete(user)
//	assert.NoError(t, err)
//
//	arr, err = db.Get[User](user)
//	assert.NoError(t, err)
//	assert.Empty(t, arr)
//	assert.True(t, arr.CreatedAt.IsZero())
//}
