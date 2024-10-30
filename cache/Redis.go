package cache

import (
	"github.com/go-redis/redis"
	"github.com/willf/bloom"
	"os"
	"strconv"
)

const (
	KeyHotQuestions     = "hot_questions"
	KeyHotQuestionTitle = "hot_questions_title"
	KeyHotAnswer        = "hot_answer"
)

type BloomFilter struct {
	filter *bloom.BloomFilter
}

// NewBloomFilter initializes a new Bloom Filter
func NewBloomFilter(size uint, hashCount float64) *BloomFilter {
	return &BloomFilter{
		filter: bloom.NewWithEstimates(size, hashCount),
	}
}

// AddUsernameToFilter adds a username to the bloom filter
func (bf *BloomFilter) AddUsernameToFilter(username string) {
	bf.filter.Add([]byte(username))
}

// CheckUsername checks if the username probably exists in the bloom filter
func (bf *BloomFilter) CheckUsername(username string) bool {
	return bf.filter.Test([]byte(username))
}

// RedisClient Redis缓存客户端单例
var RedisClient *redis.Client
var BloomF *BloomFilter

// Redis 在中间件中初始化redis链接
func Redis() {
	db, _ := strconv.ParseUint(os.Getenv("REDIS_DB"), 10, 64)
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PW"),
		DB:       int(db),
	})

	_, err := client.Ping().Result()

	if err != nil {
		panic(err)
	}

	RedisClient = client

	// Initialize Bloom Filter
	BloomF = NewBloomFilter(10000, 4)
}
