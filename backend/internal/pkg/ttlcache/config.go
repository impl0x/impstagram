package ttlcache

type cleanerType int

const (
	NoCleaner     cleanerType = iota // does not launch a cleaner, maps forever get filled and only passive cleaning works if Cache.Config.LazyDelete is true
	PerCleaner                       // Launches a goroutine per new [Cache] instance
	GlobalCleaner                    // One global goroutine which runs at the start which cleans every [Cache] ever instantiated
)

var CleaningMethod cleanerType = GlobalCleaner // Do not mutate after the server starts up, only mutate before you start the server and accept requests
