package kafka

import (
	"context"
	"testing"

	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
)

func TestKafkaAccess_getBrokerAddress(t *testing.T) {
	Convey("Test kafkaAccess.getBrokerAddress", t, func() {
		Convey("composes host:port from MQ setting", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{MQSetting: libmq.MQSetting{
				MQHost: "broker",
				MQPort: 9092,
			}}}

			So(ka.getBrokerAddress(), ShouldEqual, "broker:9092")
		})
	})
}

func TestKafkaAccess_getSASLDialer(t *testing.T) {
	Convey("Test kafkaAccess.getSASLDialer", t, func() {
		Convey("does not set SASL mechanism without auth", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{}}

			dialer := ka.getSASLDialer()
			So(dialer, ShouldNotBeNil)
			So(dialer.SASLMechanism, ShouldBeNil)
		})

		Convey("sets SASL mechanism with PLAIN auth", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{MQSetting: libmq.MQSetting{
				Auth: libmq.MQAuthSetting{
					Username:  "user",
					Password:  "pass",
					Mechanism: "PLAIN",
				},
			}}}

			dialer := ka.getSASLDialer()
			So(dialer, ShouldNotBeNil)
			So(dialer.SASLMechanism, ShouldNotBeNil)
		})

		Convey("falls back to PLAIN mechanism for unsupported mechanism", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{MQSetting: libmq.MQSetting{
				Auth: libmq.MQAuthSetting{
					Username:  "user",
					Password:  "pass",
					Mechanism: "unsupported",
				},
			}}}

			mechanism := ka.getSASLMechanism()
			So(mechanism, ShouldNotBeNil)
		})
	})
}

func TestKafkaAccess_NewWriter(t *testing.T) {
	Convey("Test kafkaAccess.NewWriter", t, func() {
		ctx := context.Background()

		Convey("creates writer without transport when auth is empty", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{MQSetting: libmq.MQSetting{
				MQHost: "broker",
				MQPort: 9092,
			}}}

			writer, err := ka.NewWriter(ctx, "topic-a")
			So(err, ShouldBeNil)
			So(writer, ShouldNotBeNil)
			So(writer.Topic, ShouldEqual, "topic-a")
			So(writer.Transport, ShouldBeNil)
		})

		Convey("creates writer with transport when auth exists", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{MQSetting: libmq.MQSetting{
				MQHost: "broker",
				MQPort: 9092,
				Auth: libmq.MQAuthSetting{
					Username:  "user",
					Password:  "pass",
					Mechanism: "SCRAM-SHA-256",
				},
			}}}

			writer, err := ka.NewWriter(ctx, "topic-b")
			So(err, ShouldBeNil)
			So(writer, ShouldNotBeNil)
			So(writer.Topic, ShouldEqual, "topic-b")
			So(writer.Transport, ShouldNotBeNil)
		})
	})
}
