package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	metrics "github.com/rcrowley/go-metrics"
)

const (
	metricPathFormat = "Kafka.%s.custom.%s.partition.%s.topic.%s.%s"
)

var (
	hostname         = os.Getenv("HOSTNAME")
	bootstrapServers = os.Getenv("KAFKA_SERVERS")
	relayHost        = os.Getenv("GRAPHITE_HOST")
	relayPort        = os.Getenv("GRAPHITE_PORT")
	prefix           = os.Getenv("GRAPHITE_PREFIX")
	retention        = 15 * time.Second
)

func main() {

	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", relayHost, relayPort))
	go graphite.Graphite(metrics.DefaultRegistry, retention, prefix, addr)

	executable := "kafka-consumer-groups"
	fmt.Println("Collecting Kafka metrics")

	for {
		cmd := exec.Command(executable,
			"--bootstrap-server",
			bootstrapServers,
			"--list")
		out, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		groups := strings.Split(string(out), "\n")
		for _, group := range groups {
			if group != "" {
				fmt.Println(group)
				describeCmd := exec.Command(executable,
					"--bootstrap-server",
					bootstrapServers,
					"--describe",
					"--group",
					group)
				desc, _ := describeCmd.Output()

				parts := strings.Split(string(desc), "\n")
				for _, line := range parts {
					if len(line) > 0 && !strings.HasPrefix(line, "TOPIC") {
						metricData := strings.Fields(line)
						if len(metricData) >= 7 {
							host := strings.Replace(metricData[6], "/", "", -1)
							host = strings.Replace(host, ".", "_", -1)
							if host != "-" {
								topic := strings.Replace(metricData[0], ".", "_", -1)
								partition := metricData[1]
								currentOffset, _ := strconv.Atoi(metricData[2])
								logEndOffset, _ := strconv.Atoi(metricData[3])
								lag, _ := strconv.Atoi(metricData[4])
								//consumerID := metricData[5]

								updateGauge("CurrentOffset", group, partition, topic, int64(currentOffset))
								updateGauge("LogEndOffset", group, partition, topic, int64(logEndOffset))
								updateGauge("Lag", group, partition, topic, int64(lag))

							}
						}
					}
				}
			}
		}
		time.Sleep(retention)
	}
}

func updateGauge(name, group, partition, topic string, value int64) {
	metricKey := fmt.Sprintf(
		metricPathFormat,
		hostname,
		group,
		partition,
		topic,
		name)
	fmt.Printf("%s => %d\n", metricKey, value)
	co := metrics.GetOrRegisterGauge(metricKey, metrics.DefaultRegistry)
	co.Update(value)
}
