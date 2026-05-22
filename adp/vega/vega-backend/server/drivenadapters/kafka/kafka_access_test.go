package kafka

import (
	"context"
	"testing"

	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/common"
)

func Test_kafkaAccess_getBrokerAddress(t *testing.T) {
	Convey("Test getBrokerAddress", t, func() {
		ka := &kafkaAccess{appSetting: &common.AppSetting{MQSetting: libmq.MQSetting{
			MQHost: "broker",
			MQPort: 9092,
		}}}

		So(ka.getBrokerAddress(), ShouldEqual, "broker:9092")
	})
}

func Test_kafkaAccess_getSASLDialer(t *testing.T) {
	Convey("Test getSASLDialer", t, func() {
		Convey("Should not set SASL mechanism without auth", func() {
			ka := &kafkaAccess{appSetting: &common.AppSetting{}}

			dialer := ka.getSASLDialer()
			So(dialer, ShouldNotBeNil)
			So(dialer.SASLMechanism, ShouldBeNil)
		})

		Convey("Should set SASL mechanism with PLAIN auth", func() {
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

		Convey("Should fall back to PLAIN mechanism for unsupported mechanism", func() {
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

func Test_kafkaAccess_NewWriter(t *testing.T) {
	Convey("Test NewWriter", t, func() {
		ctx := context.Background()

		Convey("Should create writer without transport when auth is empty", func() {
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

		Convey("Should create writer with transport when auth exists", func() {
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
