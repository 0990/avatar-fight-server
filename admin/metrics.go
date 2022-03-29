package admin

import (
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/msg/smsg"
	"github.com/0990/avatar-fight-server/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var (
	metrics_online = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "online_count",
		Help: "online player count",
	})

	metrics_rpc_millseconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        "rpc_millseconds",
		Help:        "",
		ConstLabels: nil,
		Buckets:     []float64{10, 50, 200, 500},
	}, []string{"serverId"})
)

func startMetrics() {
	serverIds := []int32{conf.GameServerID, conf.GateServerID, conf.CenterServerID}

	util.SafeGo(func() {
		for {
			time.Sleep(time.Second * 10)

			for _, v := range serverIds {
				serverId := v

				start := time.Now().UnixNano()

				metrics, err := reqMetrics(serverId)
				if err != nil {
					logrus.WithError(err).Error("req metrics")
					return
				}

				elapse := (time.Now().UnixNano() - start) / 1e6
				updateMetrics(serverId, metrics, elapse)
			}
		}
	})
}

func updateMetrics(serverId int32, metrics []*smsg.AdRespMetrics_Metrics, elapse int64) {
	updateMetricsRPC(serverId, elapse)

	for _, v := range metrics {
		switch v.Key {
		case smsg.AdRespMetrics_OnlineCount:
			updateMetricsOnline(v.Value)
		default:

		}
	}
}

func updateMetricsRPC(serverId int32, elapse int64) {
	metrics_rpc_millseconds.With(prometheus.Labels{"serverId": strconv.FormatInt(int64(serverId), 10)}).Observe(float64(elapse))
}

func updateMetricsOnline(count int32) {
	metrics_online.Set(float64(count))
}

func reqMetrics(serverId int32) ([]*smsg.AdRespMetrics_Metrics, error) {
	resp := smsg.AdRespMetrics{}
	err := Server.GetServerById(serverId).Call(&smsg.AdReqMetrics{}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Metrics, nil
}
