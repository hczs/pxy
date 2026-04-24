package detect

import (
	"fmt"
	"sort"

	"github.com/hczs/pxy/internal/config"
)

type Result struct {
	Name     string
	Path     string
	Found    bool
	Usable   bool
	Priority int
	Config   config.Config
	Reason   string
}

func PickPreferred(results []Result) (Result, bool) {
	var usable []Result
	for _, result := range results {
		if result.Found && result.Usable {
			usable = append(usable, result)
		}
	}
	if len(usable) == 0 {
		return Result{}, false
	}
	sort.SliceStable(usable, func(i, j int) bool {
		return usable[i].Priority < usable[j].Priority
	})
	return usable[0], true
}

func resultFromConfig(name, path string, priority int, cfg config.Config) Result {
	if err := cfg.Validate(); err != nil {
		return Result{Name: name, Path: path, Found: true, Priority: priority, Reason: fmt.Sprintf("invalid config: %v", err)}
	}
	return Result{Name: name, Path: path, Found: true, Usable: true, Priority: priority, Config: cfg}
}
