package model

import (
	"crypto/sha256"
	"fmt"
	"qa/cache"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

// UserLike 用户点赞表
type UserLike struct {
	gorm.Model
	UserID   uint `gorm:"not null;"` // 点赞用户Id
	AnswerID uint `gorm:"not null;"` // 被操作的回答Id
	Status   uint `gorm:"not null;"` // 点赞状态，0：无操作，1：已点赞，2：已点踩
}

const (
	NONE uint = 0
	UP   uint = 1
	DOWN uint = 2
)

const (
	AnswerLikeCount = "answer_like_count"
	UserLikeRecord  = "user_like_record"
	UserLikeAnswers = "user_like_answers"
)

func IsDeletedAnswer(aid int) bool {
	return cache.RedisClient.SIsMember(DeletedAnswers, aid).Val()
}

// DetermineTable 根据用户id返回对应的分表名
func DetermineTable(userID string, baseTableName string) string {
	hash := sha256.Sum256([]byte(userID))
	hashInt := int(hash[0])    // 取哈希值的一部分
	tableNumber := hashInt % 3 //0 1 2
	return fmt.Sprintf("%s_%d", baseTableName, tableNumber)
}

// GetUserLike 获取用户uid的点赞列表u
func GetUserLikes(uid uint) ([]uint, error) {
	var likes []UserLike
	likesTable := DetermineTable(strconv.Itoa(int(uid)), "UserLike")
	err := DB.Table(likesTable).Where("user_id = ? and status = 1", uid).
		Order("updated_at desc").Find(&likes).Error

	// 从redis中按时间倒序获取缓存的aid，
	// aid>0：说明用户对其点赞，aid<0：；说明用户对其取消点赞
	key := fmt.Sprintf("%s:%d", UserLikeAnswers, uid)
	var likesCache []int
	err = cache.RedisClient.ZRevRange(key, 0, -1).ScanSlice(&likesCache)

	var res []uint

	// set记录aid是否有更新，
	// 如果已删除则忽略，
	// 如果有更新则将数据放入res，并丢弃数据库数据
	set := make(map[uint]struct{})
	var member struct{}
	for _, aid := range likesCache {
		if IsDeletedAnswer(aid) {
			continue
		}
		if aid > 0 {
			res = append(res, uint(aid))
			set[uint(aid)] = member
		} else {
			set[uint(-aid)] = member
		}
	}

	for _, like := range likes {
		aid := like.AnswerID
		if _, ok := set[aid]; !ok {
			res = append(res, aid)
		}
	}

	return res, err
}

// GetUserLike 获取用户uid对回答aid的点赞情况
func GetUserLike(uid uint, aid uint) (uint, error) {
	// 首先从redis中获取，获取到直接返回，否则从数据库获取
	key := fmt.Sprintf("%d:%d", uid, aid)
	// 在redis中找到了
	if res, err := cache.RedisClient.HGet(UserLikeRecord, key).Result(); err == nil {
		split := strings.Split(res, ":")
		status, _ := strconv.Atoi(split[0])
		return uint(status), err
	}
	// 在redis中没有找到，从数据库获取
	var userLike UserLike
	likesTable := DetermineTable(strconv.Itoa(int(uid)), "UserLike")
	result := DB.Table(likesTable).Where("user_id = ? and answer_id = ?", uid, aid).First(&userLike)
	// 如果数据库中没有该记录，返回未点赞
	if result.RowsAffected == 0 {
		return NONE, nil
	}
	return userLike.Status, result.Error
}

// AddUserLike修改用户对某回答的点赞状态  status=0：取消点赞，status=1：点赞，status=2：点踩
func AddUserLike(uid uint, aid uint, status uint) error {
	// 如果redis中没有aid点赞数量，加载数据库中的
	err := cache.RedisClient.HGet(AnswerLikeCount, strconv.Itoa(int(aid))).Err()
	if err == redis.Nil {
		err = nil
		ans, err := GetAnswer(aid)
		cnt := ans.LikeCount
		err = cache.RedisClient.HSet(AnswerLikeCount, strconv.Itoa(int(aid)), cnt).Err()
		if err != nil {
			return err
		}
	}
	// 获取之前的点赞状态
	pre, err := GetUserLike(uid, aid)
	if err != nil {
		return err
	}
	var incr int64 = 0
	pipe := cache.RedisClient.TxPipeline()
	// 根据前后的状态，判断点赞数的增减与否
	keyAns := fmt.Sprintf("%s:%d", UserLikeAnswers, uid)
	if (pre == NONE || pre == DOWN) && status == UP {
		incr = 1
		pipe.ZRem(keyAns, -int(aid))
		pipe.ZAdd(keyAns, redis.Z{Score: float64(time.Now().Unix()), Member: aid})
	} else if pre == UP && (status == NONE || status == DOWN) {
		incr = -1
		pipe.ZRem(keyAns, aid)
		pipe.ZAdd(keyAns, redis.Z{Score: float64(time.Now().Unix()), Member: -int(aid)})
	}
	pipe.HIncrBy(AnswerLikeCount, strconv.Itoa(int(aid)), incr)
	keyRec := fmt.Sprintf("%d:%d", uid, aid)
	pipe.HSet(UserLikeRecord, keyRec,
		fmt.Sprintf("%d:%d", status, time.Now().Unix()))
	_, err = pipe.Exec()
	return err
}

// GetLikeCountIdInCache 根据AnswerID获取缓存中点赞数据是否存在与修改总数
func GetLikeCountInCache(aid uint) (bool, uint, error) {
	res, err := cache.RedisClient.HGet(AnswerLikeCount, strconv.Itoa(int(aid))).Int()
	if err == redis.Nil {
		return false, 0, nil
	}
	return true, uint(res), err
}

// SyncUserLikeRecord 将redis中的用户点赞记录同步到数据库，对应like表
func SyncUserLikeRecord() {
	fmt.Println("Start SyncUserLikeRecord...")
	defer fmt.Println("End SyncUserLikeRecord...")

	// 从redis中获得数据
	data := cache.RedisClient.HGetAll(UserLikeRecord).Val()

	for key, val := range data {
		//fmt.Printf("%s\t%s\n", key, val)

		splitK := strings.Split(key, ":")
		uid, _ := strconv.Atoi(splitK[0])
		aid, _ := strconv.Atoi(splitK[1])

		// 回答已删除则不更新
		if IsDeletedAnswer(aid) {
			continue
		}

		splitV := strings.Split(val, ":")
		status, _ := strconv.Atoi(splitV[0])
		updateTime, _ := strconv.ParseInt(splitV[1], 10, 64)

		var userLike UserLike
		userLike.UserID = uint(uid)
		userLike.AnswerID = uint(aid)

		likesTable := DetermineTable(strconv.Itoa(int(uid)), "UserLike")
		row := DB.Table(likesTable).Where(&userLike).Find(&userLike).RowsAffected

		userLike.UpdatedAt = time.Unix(updateTime, 0)

		var err error
		// 存在则更新，不存在则创建
		if row > 0 {
			userLike.Status = uint(status)
			likesTable := DetermineTable(strconv.Itoa(int(uid)), "UserLike")
			err = DB.Table(likesTable).Save(&userLike).Error
		} else {
			userLike.Status = uint(status)
			likesTable := DetermineTable(strconv.Itoa(int(uid)), "UserLike")
			err = DB.Table(likesTable).Create(&userLike).Error
		}
		if err != nil {
			panic(err)
		}
	}

	// 删除redis中的数据
	cache.RedisClient.Del(UserLikeRecord)

	// 匹配所有的UserLikeAnswers相关的key，将其删除
	var keys []string
	cache.RedisClient.Keys(fmt.Sprintf("%s*", UserLikeAnswers)).ScanSlice(&keys)
	for _, key := range keys {
		cache.RedisClient.Del(key)
	}
}

// SyncAnswerLikeCount 将redis中的回答点赞数量同步到数据库，对应answer表
func SyncAnswerLikeCount() {
	fmt.Println("Start SyncLikeCount...")
	defer fmt.Println("End SyncLikeCount...")

	// 从redis中获得数据
	data := cache.RedisClient.HGetAll(AnswerLikeCount).Val()

	tx := DB.Begin()
	for key, val := range data {
		//fmt.Printf("%s\t%s\n", key, val)

		id, _ := strconv.Atoi(key)
		count, _ := strconv.Atoi(val)

		// 回答已删除则不更新
		if IsDeletedAnswer(id) {
			continue
		}

		var answer Answer
		answer.ID = uint(id)

		err := DB.Model(&answer).Update("like_count", count).Error
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	tx.Commit()
	// 删除redis中的数据
	cache.RedisClient.Del(AnswerLikeCount)
}

func FreeDeletedAnswersRecord() {
	cache.RedisClient.Del(DeletedAnswers)
}
