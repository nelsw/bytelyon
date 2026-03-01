package test

//func Test_Client_Dynamo_Item_Functions(t *testing.T) {
//
//	var err error
//	ctx := context.Background()
//	c := util.Must(dbClient.New())
//
//	exp := &model.User{util.Must(uuid.NewV7())}
//
//	err = dbClient.PutItem(ctx, c, exp)
//	assert.NoError(t, err)
//
//	var act = model.User{exp.ID}
//	_, err = dbClient.GetItem[model.User](ctx, c, exp)
//	assert.NoError(t, err)
//	assert.Equal(t, exp.ID, act.ID)
//
//	err = dbClient.DeleteItem(ctx, c, exp)
//	assert.NoError(t, err)
//
//	act, err = dbClient.GetItem[model.User](ctx, c, exp)
//	assert.ErrorAs(t, err, &dbClient.NotFoundEx)
//}
//
//func Test_Client_Dynamo_Query(t *testing.T) {
//
//	userIDs := []uuid.UUID{util.Must(uuid.NewV7()), util.Must(uuid.NewV7())}
//
//	var arr []dbClient.Entity
//
//	for i := 0; i < 10; i++ {
//
//		bot := model.Bot{
//			UserID:    userIDs[i%2],
//			BotID:     util.Must(uuid.NewV7()),
//			BlackList: []string{fake.DomainName()},
//			Frequency: time.Hour * time.Duration(fake.Uint8()),
//			Target:    fake.URL(),
//			UpdatedAt: time.Now(),
//		}
//
//		if i < 3 {
//			arr = append(arr, &model.SearchBot{Bot: bot, Headless: false})
//		} else if i < 5 {
//			arr = append(arr, &model.SitemapBot{Bot: bot})
//		} else {
//			arr = append(arr, &model.NewsBot{Bot: bot})
//		}
//	}
//
//	var err error
//	ctx := context.Background()
//	dbc := util.Must(dbClient.New())
//
//	for _, e := range arr {
//		assert.NoError(t, dbClient.PutItem(ctx, dbc, e))
//	}
//
//	var bots []model.SearchBot
//	bot := &model.SearchBot{}
//	bots, err = dbClient.QueryByID[model.SearchBot](ctx, dbc, bot.Name(), "UserID", userIDs[0])
//
//	assert.NoError(t, err)
//	assert.NotEmpty(t, bots)
//
//	util.PrettyPrintln(bots)
//}
//
//func Test_Client_Dynamo_Email(t *testing.T) {
//	ctx := context.Background()
//	c := util.Must(dbClient.New())
//
//	email, err := dbClient.GetItem[model.Email](ctx, c, &model.Email{ID: "kowalski7012@gmail.com"})
//	assert.NoError(t, err)
//	assert.Equal(t, "kowalski7012@gmail.com", email.ID)
//
//	var user model.User
//	user, err = dbClient.GetItem[model.User](ctx, c, &model.User{ID: email.UserID})
//	assert.NoError(t, err)
//	assert.Equal(t, email.UserID, user.ID)
//}
//
//func Test_Client_Dynamo_Token(t *testing.T) {
//	ctx := context.Background()
//	c := util.Must(dbClient.New())
//
//	//exp := &model.Token{
//	//	ID:     util.Must(uuid.NewV7()),
//	//	UserID: util.Must(uuid.NewV7()),
//	//	Type:   model.ConfirmEmailTokenType,
//	//	Expiry: time.Now().Add(time.Hour),
//	//}
//	//assert.NoError(t, dbClient.DeleteTable(ctx, c, exp))
//	//assert.NoError(t, dbClient.CreateTable(ctx, c, exp))
//	//assert.NoError(t, dbClient.PutItem(ctx, c, exp))
//
//	act, err := dbClient.GetItem[model.Token](ctx, c, &model.Token{})
//	assert.NoError(t, err)
//	assert.Equal(t, act.ID, uuid.Nil)
//	//assert.Equal(t, exp.ID, act.ID)
//	//assert.Equal(t, exp.Type, act.Type)
//	//assert.Equal(t, exp.Expiry.Truncate(60*time.Second), act.Expiry.Truncate(60*time.Second))
//	//assert.Equal(t, exp.UserID, act.UserID)
//}
//
//func Test_Client_Dynamo_Find_Users(t *testing.T) {
//	c := util.Must(dbClient.New())
//
//	arr, _ := dbClient.Scan[model.User](c, util.Ptr(model.User{}).Name())
//
//	//ctx := context.Background()
//	//c := util.Must(dbClient.New())
//	//arr, err := getBots(ctx, c, model.User{ID: uuid.MustParse("019ca9f3-239f-7908-8c79-27ead82bdca5")}, &model.SearchBot{})
//	//assert.NoError(t, err)
//	//fmt.Println(len(arr))
//	//util.PrettyPrintln(arr)
//	//
//	//arr2, err2 := getBots(ctx, c, model.User{ID: uuid.MustParse("019ca9f3-239f-7908-8c79-27ead82bdca5")}, &model.SitemapBot{})
//	//assert.NoError(t, err)
//	fmt.Println(len(arr))
//	util.PrettyPrintln(arr)
//	//
//	//var arr3 []any
//	//arr3 = append(arr3, arr...)
//}
