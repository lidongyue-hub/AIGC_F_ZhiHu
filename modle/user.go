package model

import (
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Username    string      `gorm:"unique;not null;"`                              // 用户名
	Password    string      `gorm:"not null;"`                                     // 密码
	UserProfile UserProfile `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // 关联用户信息
	Questions   []Question  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // 关联问题信息
	Answers     []Answer    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // 关联回答信息
}

// UserProfile 用户信息模型
type UserProfile struct {
	gorm.Model
	UserID      uint
	Nickname    string `gorm:"default:null"`         // 昵称
	Email       string `gorm:"unique;default:null;"` // 邮箱
	Avatar      string `gorm:"default:null;"`        // 头像
	Status      int    `gorm:"not null;default:0;"`  // 状态
	Description string `gorm:"default:null"`         // 个人描述
}

const (
	// PasswordCost 密码加密难度
	PasswordCost = bcrypt.DefaultCost
	// Inactive 未激活用户
	Inactive int = 0
	// Active 激活用户
	Active int = 1
)

// DetermineTable 根据用户名返回对应的分表名
func DetermineTable(username string, baseTableName string) string {
	hash := sha256.Sum256([]byte(username))
	hashInt := int(hash[0]) // 取哈希值的一部分
	tableNumber := hashInt % 3
	return fmt.Sprintf("%s_%d", baseTableName, tableNumber)
}

// BeforeCreate是一个GORM钩子，在将新的用户记录插入数据库之前执行。
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// 如果UserProfile不为空，手动将其插入到相应的表中
	if u.UserProfile != (UserProfile{}) {
		u.UserProfile.UserID = u.ID // Ensure that the UserID is set correctly

		// 确定UserProfile使用哪个表
		profileTable := DetermineTable(u.Username, "UserProfile") // Use a unique field to determine the table

		// 将UserProfile插入到确定的表中
		if err := tx.Table(profileTable).Create(&u.UserProfile).Error; err != nil {
			return err
		}

		// Prevent GORM from automatically creating a UserProfile record
		u.UserProfile = UserProfile{}
	}

	return nil
}

// GetUser 用ID获取用户
func GetUser(ID interface{}) (User, error) {
	var user User
	result := DB.First(&user, ID)
	return user, result.Error
}

// GetUserProfile 用ID获取用户详细信息
func GetUserProfile(ID interface{}) (UserProfile, error) { //根据ID查User查到用户(名)，再根据用户(名)查个人信息表 || 直接根据ID查UserProfile，但需要遍历这三个表
	var user User
	_ = DB.First(&user, ID)

	var profile UserProfile
	profileTable := DetermineTable(user.Username, "UserProfile")
	result := DB.Table(profileTable).Where("nickname = ?", user.Username).First(&profile) // user_id
	return profile, result.Error                                                          // user.UserProfile, result.Error
}

// SetPassword 设置密码
func (user *User) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordCost)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

// CheckPassword 校验密码
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}
