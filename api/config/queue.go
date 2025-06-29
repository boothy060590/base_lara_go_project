package config

import "base_lara_go_project/app/core/env"

// QueueConfig returns the queue configuration with fallback values
func QueueConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.GetEnv("QUEUE_CONNECTION", "sync"),
		"connections": map[string]interface{}{
			"sync": map[string]interface{}{
				"driver": "sync",
			},
			"sqs": map[string]interface{}{
				"driver":   "sqs",
				"key":      env.GetEnv("SQS_ACCESS_KEY", "local"),
				"secret":   env.GetEnv("SQS_SECRET_KEY", "local"),
				"region":   env.GetEnv("SQS_REGION", "us-east-1"),
				"queue":    env.GetEnv("SQS_QUEUE", "default"),
				"endpoint": env.GetEnv("SQS_ENDPOINT", "http://localhost:9324"),
			},
		},
		"queues": map[string]interface{}{
			"jobs":   env.GetEnv("SQS_QUEUE_JOBS", "default"),
			"mail":   env.GetEnv("SQS_QUEUE_MAIL", "default"),
			"events": env.GetEnv("SQS_QUEUE_EVENTS", "default"),
		},
		"enabled_queues": []string{
			env.GetEnv("SQS_QUEUE_JOBS", "default"),
			env.GetEnv("SQS_QUEUE_MAIL", "default"),
			env.GetEnv("SQS_QUEUE_EVENTS", "default"),
		},
		"workers": map[string]interface{}{
			"default": map[string]interface{}{
				"queues":       []string{"mail", "jobs", "events", "default"},
				"max_jobs":     env.GetEnvInt("WORKER_MAX_JOBS", 1000),
				"memory_limit": env.GetEnvInt("WORKER_MEMORY_LIMIT", 128),
				"timeout":      env.GetEnvInt("WORKER_TIMEOUT", 60),
				"sleep":        env.GetEnvInt("WORKER_SLEEP", 3),
				"tries":        env.GetEnvInt("WORKER_TRIES", 3),
			},
		},
		"api_queues": map[string]interface{}{
			"mail":    "mail",
			"jobs":    "jobs",
			"events":  "events",
			"default": "default",
		},
	}
}
