package nsqconsumer

import (
	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/config"
)

// Run runs the specified AuthzServer. This should never exit.
func Run(cfg *config.Config) error {
	manager, err := createConsumerManager(cfg)
	if err != nil {
		return err
	}
	return manager.Run()
}
