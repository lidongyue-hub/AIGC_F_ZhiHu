package model

//执行数据迁移

func migration() {
	// 自动迁移模式
	_ = DB.AutoMigrate(&User{}, &Question{}, &Answer{}, &UserLike{})
	//_ = DB.Table("User_1").AutoMigrate(&User{})
	//_ = DB.Table("User_2").AutoMigrate(&User{})
	//_ = DB.Table("User_3").AutoMigrate(&User{})
	_ = DB.Table("UserProfile_1").AutoMigrate(&UserProfile{})
	_ = DB.Table("UserProfile_2").AutoMigrate(&UserProfile{})
	_ = DB.Table("UserProfile_3").AutoMigrate(&UserProfile{})

}
