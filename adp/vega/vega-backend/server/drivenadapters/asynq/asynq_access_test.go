package asynq

import (
	"testing"

	asynqlib "github.com/hibiken/asynq"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
)

func TestAsynqAccess_getRedisClientOpt(t *testing.T) {
	Convey("Test asynqAccess.getRedisClientOpt", t, func() {
		Convey("uses standalone redis address", func() {
			aqa := &asynqAccess{appSetting: &common.AppSetting{RedisSetting: common.RedisSetting{
				ConnectType: "standalone",
				Host:        "redis",
				Port:        6379,
				Username:    "user",
				Password:    "pass",
			}}}

			opt := aqa.getRedisClientOpt()
			clientOpt, ok := opt.(asynqlib.RedisClientOpt)
			So(ok, ShouldBeTrue)
			So(clientOpt.Addr, ShouldEqual, "redis:6379")
			So(clientOpt.Username, ShouldEqual, "user")
			So(clientOpt.Password, ShouldEqual, "pass")
		})

		Convey("uses cluster setting as redis client address", func() {
			aqa := &asynqAccess{appSetting: &common.AppSetting{RedisSetting: common.RedisSetting{
				ConnectType: "cluster",
				Host:        "cluster-redis",
				Port:        6380,
			}}}

			opt := aqa.getRedisClientOpt()
			clientOpt, ok := opt.(asynqlib.RedisClientOpt)
			So(ok, ShouldBeTrue)
			So(clientOpt.Addr, ShouldEqual, "cluster-redis:6380")
		})

		Convey("uses master address in master-slave mode", func() {
			aqa := &asynqAccess{appSetting: &common.AppSetting{RedisSetting: common.RedisSetting{
				ConnectType: "master-slave",
				MasterHost:  "master",
				MasterPort:  6381,
			}}}

			opt := aqa.getRedisClientOpt()
			clientOpt, ok := opt.(asynqlib.RedisClientOpt)
			So(ok, ShouldBeTrue)
			So(clientOpt.Addr, ShouldEqual, "master:6381")
		})

		Convey("uses sentinel failover options", func() {
			aqa := &asynqAccess{appSetting: &common.AppSetting{RedisSetting: common.RedisSetting{
				ConnectType:      "sentinel",
				Username:         "redis-user",
				Password:         "redis-pass",
				SentinelHost:     "sentinel",
				SentinelPort:     26379,
				SentinelUsername: "sentinel-user",
				SentinelPassword: "sentinel-pass",
				MasterGroupName:  "mymaster",
			}}}

			opt := aqa.getRedisClientOpt()
			failoverOpt, ok := opt.(asynqlib.RedisFailoverClientOpt)
			So(ok, ShouldBeTrue)
			So(failoverOpt.SentinelAddrs, ShouldResemble, []string{"sentinel:26379"})
			So(failoverOpt.Username, ShouldEqual, "redis-user")
			So(failoverOpt.Password, ShouldEqual, "redis-pass")
			So(failoverOpt.SentinelUsername, ShouldEqual, "sentinel-user")
			So(failoverOpt.SentinelPassword, ShouldEqual, "sentinel-pass")
			So(failoverOpt.MasterName, ShouldEqual, "mymaster")
		})

		Convey("falls back to standalone options for unknown connect type", func() {
			aqa := &asynqAccess{appSetting: &common.AppSetting{RedisSetting: common.RedisSetting{
				ConnectType: "unknown",
				Host:        "fallback",
				Port:        6379,
			}}}

			opt := aqa.getRedisClientOpt()
			clientOpt, ok := opt.(asynqlib.RedisClientOpt)
			So(ok, ShouldBeTrue)
			So(clientOpt.Addr, ShouldEqual, "fallback:6379")
		})
	})
}
