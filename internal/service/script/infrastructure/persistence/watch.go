package persistence

import (
	"context"
	"strconv"

	goRedis "github.com/go-redis/redis/v8"
	"github.com/scriptscat/scriptlist/internal/service/script/domain/repository"
	"github.com/scriptscat/scriptlist/pkg/utils"
	"gorm.io/gorm"
)

type watch struct {
	db    *gorm.DB
	redis *goRedis.Client
}

func NewScriptWatch(db *gorm.DB, redis *goRedis.Client) repository.ScriptWatch {
	return &watch{db: db, redis: redis}
}

func (w *watch) key(issue int64) string {
	return "script:watch:" + strconv.FormatInt(issue, 10)
}

func (w *watch) List(script int64) ([]*repository.Watch, error) {
	list, err := w.redis.HGetAll(context.Background(), w.key(script)).Result()
	if err != nil {
		return nil, err
	}
	ret := make([]*repository.Watch, 0)
	for k, v := range list {
		ret = append(ret, &repository.Watch{UserId: utils.StringToInt64(k), Level: utils.StringToInt(v)})
	}
	return ret, nil
}

func (w *watch) Num(script int64) (int, error) {
	list, err := w.redis.HGetAll(context.Background(), w.key(script)).Result()
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

func (w *watch) Watch(script, user int64, level int) error {
	return w.redis.HSet(context.Background(), w.key(script), user, level).Err()
}

func (w *watch) Unwatch(script, user int64) error {
	return w.redis.HDel(context.Background(), w.key(script), strconv.FormatInt(user, 10)).Err()
}

func (w *watch) IsWatch(script, user int64) (int, error) {
	ret, err := w.redis.HGet(context.Background(), w.key(script), strconv.FormatInt(user, 10)).Result()
	if err != nil {
		if err == goRedis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return utils.StringToInt(ret), nil
}
