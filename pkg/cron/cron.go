package cron

import (
	"time"

	"github.com/busy-cloud/boat/pool"
	"github.com/go-co-op/gocron/v2"
)

var scheduler gocron.Scheduler

func init() {
	var err error
	scheduler, err = gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
}

func Interval(interval int64, fn func()) (*Job, error) {
	job, err := scheduler.NewJob(
		gocron.DurationJob(time.Second*time.Duration(interval)),
		gocron.NewTask(func() { _ = pool.Insert(fn) }),
	)
	return &Job{job}, err
}

func Crontab(crontab string, fn func()) (*Job, error) {
	job, err := scheduler.NewJob(
		gocron.CronJob(crontab, false),
		gocron.NewTask(func() { _ = pool.Insert(fn) }),
	)
	return &Job{job}, err
}
