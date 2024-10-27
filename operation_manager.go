package main

import (
	"fmt"
	"sync"
)

type OperationManager struct {
	retryConfigs map[RequestType][]RetryConfig
	mu           sync.Mutex
}

func NewOperationManager(config Config) *OperationManager {
	rc := make(map[RequestType][]RetryConfig)
	om := &OperationManager{
		retryConfigs: rc,
	}
	for _, retryConfig := range config.RetryConfig {
		om.addRetryConfig(retryConfig)
	}
	fmt.Printf("%+v\n", om)
	return om
}

// Empty string represent there is no plantation required.
func (om *OperationManager) retrieveOperation(requestType RequestType) string {
	om.mu.Lock()
	defer om.mu.Unlock()

	configs, ok := om.retryConfigs[requestType]
	if !ok {
		return ""
	}

	for len(configs) > 0 {
		cc := &configs[0]
		if cc.SkipCount > 0 {
			cc.SkipCount--
			return ""
		} else if cc.RetryCount > 0 {
			cc.RetryCount--
			return cc.RetryInstruction
		} else {
			configs = configs[1:]
			om.retryConfigs[requestType] = configs
		}
	}
	return ""
}

func (om *OperationManager) addRetryConfig(rc RetryConfig) {
	rt := RequestType(rc.Method)
	println(rt)
	if _, ok := om.retryConfigs[rt]; ok {
		// Key exists, append the new retryConfig to the existing list
		om.retryConfigs[rt] = append(om.retryConfigs[rt], rc)
	} else {
		// Key doesn't exist, create a new list with the retryConfig
		om.retryConfigs[rt] = []RetryConfig{rc}
	}
}
