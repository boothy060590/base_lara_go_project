package config

func QueueConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": getEnv("QUEUE_CONNECTION", "sqs"),
		"connections": map[string]interface{}{
			"sqs": map[string]interface{}{
				"driver":   "sqs",
				"key":      getEnv("SQS_ACCESS_KEY", "local"),
				"secret":   getEnv("SQS_SECRET_KEY", "local"),
				"region":   getEnv("SQS_REGION", "us-east-1"),
				"queue":    getEnv("SQS_QUEUE", "default"),
				"endpoint": getEnv("SQS_ENDPOINT", "http://localhost:9324"),
			},
		},
		"queues": map[string]interface{}{
			"jobs":   getEnv("SQS_QUEUE_JOBS", "default"),
			"mail":   getEnv("SQS_QUEUE_MAIL", "default"),
			"events": getEnv("SQS_QUEUE_EVENTS", "default"),
		},
		"enabled_queues": []string{
			getEnv("SQS_QUEUE_JOBS", "default"),
			getEnv("SQS_QUEUE_MAIL", "default"),
			getEnv("SQS_QUEUE_EVENTS", "default"),
		},
	}
}
