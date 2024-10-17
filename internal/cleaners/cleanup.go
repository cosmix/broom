package cleaners

import (
	"fmt"
	"sort"
	"strings"
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

func GetAllCleanupTypes() []string {
	var types []string
	cleanupFunctions.Range(func(key, value interface{}) bool {
		types = append(types, key.(string))
		return true
	})
	sort.Strings(types)
	return types
}

func PerformCleanup(cleanupType string) (uint64, error) {
	startSpace := utils.GetFreeDiskSpace()

	cleanerInterface, ok := cleanupFunctions.Load(cleanupType)
	if !ok {
		return 0, fmt.Errorf("unknown cleanup type: %s", cleanupType)
	}

	cleaner := cleanerInterface.(Cleaner)
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

func GetCleanupDescription() string {
	var sb strings.Builder
	sb.WriteString("Available cleanup types:\n")

	types := GetAllCleanupTypes()
	for _, t := range types {
		sb.WriteString(fmt.Sprintf("- %s\n", t))
	}

	return sb.String()
}

func CleanupRequiresConfirmation(cleanupType string) bool {
	cleanerInterface, ok := cleanupFunctions.Load(cleanupType)
	if !ok {
		return false
	}
	cleaner := cleanerInterface.(Cleaner)
	return cleaner.RequiresConfirmation
}
