package cron

import "github.com/go-co-op/gocron/v2"

type Job struct {
	gocron.Job
}

func (j *Job) Stop() error {
	return scheduler.RemoveJob(j.ID())
}
