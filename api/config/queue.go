package config

import "base_lara_go_project/app/core/laravel_core/env"

// QueueConfig returns the queue configuration with fallback values
func QueueConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("QUEUE_CONNECTION", "sync"),
		"connections": map[string]interface{}{
			"sync": map[string]interface{}{
				"driver": "sync",
			},
			"sqs": map[string]interface{}{
				"driver":   "sqs",
				"key":      env.Get("SQS_ACCESS_KEY", "local"),
				"secret":   env.Get("SQS_SECRET_KEY", "local"),
				"region":   env.Get("SQS_REGION", "us-east-1"),
				"queue":    env.Get("SQS_QUEUE", "default"),
				"endpoint": env.Get("SQS_ENDPOINT", "http://localhost:9324"),
			},
		},
		"queues": map[string]interface{}{
			"jobs":   env.Get("SQS_QUEUE_JOBS", "default"),
			"mail":   env.Get("SQS_QUEUE_MAIL", "default"),
			"events": env.Get("SQS_QUEUE_EVENTS", "default"),
		},
		"enabled_queues": []string{
			env.Get("SQS_QUEUE_JOBS", "default"),
			env.Get("SQS_QUEUE_MAIL", "default"),
			env.Get("SQS_QUEUE_EVENTS", "default"),
		},
		"workers": map[string]interface{}{
			"default": map[string]interface{}{
				"queues":       []string{"mail", "jobs", "events", "default"},
				"max_jobs":     env.GetInt("WORKER_MAX_JOBS", 1000),
				"memory_limit": env.GetInt("WORKER_MEMORY_LIMIT", 128),
				"timeout":      env.GetInt("WORKER_TIMEOUT", 60),
				"sleep":        env.GetInt("WORKER_SLEEP", 3),
				"tries":        env.GetInt("WORKER_TRIES", 3),
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
