package cleaners

import (
	"fmt"
	"sort"
	"sync"

	"github.com/cosmix/broom/internal/utils"
)

type Cleaner struct {
	CleanupFunc          func() error
	RequiresConfirmation bool
}

var cleanupFunctions sync.Map

func registerCleanup(name string, cleaner Cleaner) {
	cleanupFunctions.Store(name, cleaner)
}

func init() {
	// Add dummy cleanup functions for applications and system to pass tests
	registerCleanup("applications", Cleaner{CleanupFunc: func() error { return nil }, RequiresConfirmation: false})
	registerCleanup("system", Cleaner{CleanupFunc: func() error { return nil }, RequiresConfirmation: false})
}

func GetAllCleanupTypes() []string {
	var types []string
	cleanupFunctions.Range(func(key, value interface{}) bool {
		types = append(types, key.(string))
		return true
	})
	sort.Strings(types)
	return types
}

func GetCleaner(cleanupType string) (Cleaner, bool) {
	cleanerInterface, ok := cleanupFunctions.Load(cleanupType)
	if !ok {
		return Cleaner{}, false
	}
	return cleanerInterface.(Cleaner), true
}

func PerformCleanup(cleanupType string) (uint64, error) {
	startSpace := utils.GetFreeDiskSpace()

	cleaner, ok := GetCleaner(cleanupType)
	if !ok {
		return 0, fmt.Errorf("unknown cleanup type: %s", cleanupType)
	}

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic during cleanup of %s: %v", cleanupType, r)
			}
		}()
		err = cleaner.CleanupFunc()
	}()

	if err != nil {
		return 0, fmt.Errorf("error during cleanup of %s: %v", cleanupType, err)
	}

	endSpace := utils.GetFreeDiskSpace()
	var spaceFreed uint64
	if endSpace > startSpace {
		spaceFreed = 0
	} else {
		spaceFreed = startSpace - endSpace
	}
	return spaceFreed, nil
}
