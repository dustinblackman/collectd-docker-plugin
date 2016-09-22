package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/camelcase"
	"github.com/fatih/structs"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/urfave/cli"
)

var (
	host    = "localhost"
	version = "HEAD"
)

func printCollectD(containerName, statType, statTypeInstance string, value uint64) {
	valueString := strconv.FormatUint(value, 10)
	fmt.Printf("PUTVAL \"%s/docker-%s/%s-%s\" interval=60 N:%s\n", host, containerName, statType, statTypeInstance, valueString)
}

func toUnderscore(key string) string {
	return strings.ToLower(strings.Join(camelcase.Split(key), "_"))
}

func getStats(client *docker.Client, container *docker.Container, waitTime int) {
	containerName := container.Name[1:len(container.Name)]
	errC := make(chan error, 1)
	statsC := make(chan *docker.Stats)

	go func() {
		errC <- client.Stats(docker.StatsOptions{ID: container.ID, Stats: statsC, Stream: true})
	}()

	currentInterval := 0
	for {
		stats, ok := <-statsC
		if !ok {
			spew.Dump(stats)
			break
		}

		// Wait time
		currentInterval++
		if currentInterval != waitTime {
			continue
		}
		currentInterval = 0

		// Memory
		printCollectD(containerName, "memory", "max_usage", stats.MemoryStats.MaxUsage)
		printCollectD(containerName, "memory", "usage", stats.MemoryStats.Usage)
		for key, value := range structs.Map(stats.MemoryStats.Stats) {
			printCollectD(containerName, "memory", toUnderscore(key), value.(uint64))
		}

		// CPU
		printCollectD(containerName, "cpu", "total_usage", stats.CPUStats.CPUUsage.TotalUsage)

		// Network
		mergedNetworks := map[string]uint64{}
		for _, value := range stats.Networks {
			for networkKey, networkValue := range structs.Map(value) {
				if _, ok := mergedNetworks[networkKey]; ok {
					mergedNetworks[networkKey] += networkValue.(uint64)
				} else {
					mergedNetworks[networkKey] = networkValue.(uint64)
				}
			}
		}
		for key, value := range mergedNetworks {
			printCollectD(containerName, "network", toUnderscore(key), value)
		}
	}

	err := <-errC
	if err != nil {
		log.Fatal(err)
	}
}

func listContainers(ctx *cli.Context) {
	host = ctx.String("collectd-hostname")

	var client *docker.Client
	var err interface{}
	if ctx.Bool("docker-environment") {
		client, err = docker.NewClientFromEnv()
	} else {
		client, err = docker.NewClient(ctx.String("docker-host"))
	}

	if err != nil {
		log.Fatal(err)
		return
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: false, Size: false})
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, container := range containers {
		containerDetails, err := client.InspectContainer(container.ID)
		if err != nil {
			log.Fatal(err)
		}
		go getStats(client, containerDetails, ctx.Int("wait-time"))
	}

	go func() {
		dockerEvents := make(chan *docker.APIEvents, 100)
		client.AddEventListener(dockerEvents)
		for event := range dockerEvents {
			if event.Status == "start" {
				containerDetails, err := client.InspectContainer(event.ID)
				if err != nil {
					log.Fatal(err)
				}
				go getStats(client, containerDetails, ctx.Int("wait-time"))
			}
		}
	}()

	// Wait forever.
	select {}
}

func main() {
	app := cli.NewApp()
	app.Name = "collectd-docker-plugin"
	app.Usage = "A collectd plugin to submit metrics from the docker stats API"
	app.Version = version
	app.Author = "Dustin Blackman"
	app.Copyright = "(c) 2016 " + app.Author
	app.EnableBashCompletion = true
	app.Action = listContainers

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker-host, d",
			Usage: "Docker host",
			Value: "unix:///var/run/docker.sock",
		},
		cli.BoolFlag{
			Name:  "docker-environment, de",
			Usage: "Use environment docker variables instead of passing docker socket path",
		},
		cli.StringFlag{
			Name:   "collectd-hostname, ch",
			Usage:  "Docker host",
			EnvVar: "COLLECTD_HOSTNAME",
			Value:  "localhost",
		},
		cli.IntFlag{
			Name:  "wait-time, w",
			Usage: "Wait time between how often stats should be submitted to collectd",
			Value: 3,
		},
	}

	app.Run(os.Args)
}
