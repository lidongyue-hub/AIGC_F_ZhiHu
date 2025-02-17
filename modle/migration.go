package model

//执行数据迁移

func migration() {
	// 自动迁移模式
	_ = DB.AutoMigrate(&User{}, &Question{}, &Answer{}, UserProfile{})

	_ = DB.Table("UserLike_0").AutoMigrate(&UserLike{})
	_ = DB.Table("UserLike_1").AutoMigrate(&UserLike{})
	_ = DB.Table("UserLike_2").AutoMigrate(&UserLike{})
}
