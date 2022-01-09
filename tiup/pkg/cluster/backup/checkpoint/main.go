package main

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/pingcap/tiup/pkg/cluster/api"
	"github.com/spf13/pflag"
)

var (
	cluster            = pflag.String("cluster-id", "", "the cluster for updating")
	authKey            = pflag.String("auth-key", "", "the authkey of your account")
	checkpointInterval = pflag.Duration("checkpoint-interval", 60*time.Second, "the interval of creating checkpoints")
	url                = pflag.String("url", "s3://pcloud2021/backups", "the url")
)

func run(ctx context.Context, timer <-chan time.Time) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer:
			clusterInfo, err := api.GetCluster(*cluster, *authKey)
			if err != nil {
				return err
			}
			if clusterInfo.Cluster.SetupStatus != "finish" {
				continue
			}
			cp, err := api.CreateCheckpoint(api.CreateCheckpointRequest{
				AuthKey:        *authKey,
				ClusterID:      *cluster,
				UploadStatus:   "finish",
				UploadProgress: 100,
				CheckpointTime: time.Now().UnixMilli(),
				URL:            *url,
				// How can we calcute the backup size?
				// Maybe we must inject the logic into CDC?
				// (Maybe it is a little dirty to let CDC known about the "cloud" API?)
				BackupSize: 42,
				Operator:   "pingcap",
			})
			if err != nil {
				return err
			}
			fmt.Println(color.GreenString("Checkpoint %s created.", cp))
		}
	}
}

func redirect(file string) error {
	if err := syscall.Close(1); err != nil {
		return err
	}
	fd, err := syscall.Open(file, syscall.O_APPEND|syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	if fd != 1 {
		return errors.New("failed to redirect stdout")
	}
	return nil
}

func main() {
	pflag.Parse()
	if err := redirect("./log.txt"); err != nil {
		panic(err)
	}
	fmt.Println("welcome to the hacky checkpoint uploader.")
	tick := time.NewTicker(*checkpointInterval)
	if err := run(context.Background(), tick.C); err != nil {
		panic(err)
	}
}
