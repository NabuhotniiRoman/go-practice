package build

var (
	// Version додатку (встановлюється через ldflags)
	Version = "dev"

	// Number білда (встановлюється через ldflags)
	Number = "local"

	// GitCommit хеш коміту (встановлюється через ldflags)
	GitCommit = "unknown"

	// BuildTime час збірки (встановлюється через ldflags)
	BuildTime = "unknown"
)

// Info повертає інформацію про білд
func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"number":     Number,
		"git_commit": GitCommit,
		"build_time": BuildTime,
	}
}
