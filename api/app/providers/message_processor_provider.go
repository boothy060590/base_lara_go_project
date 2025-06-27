package providers

import (
	"base_lara_go_project/app/core"
)

func RegisterMessageProcessor() {
	// Create message processor provider and set global instance
	messageProcessorProvider := core.NewMessageProcessorProvider()
	core.SetMessageProcessorService(messageProcessorProvider)
}
