package model

const (
	CPU_TIME        = "cpu_time"
	NETWORK_TIME    = "network_time"
	LOCK_GC_TINME   = "lock_gc_time"
	DISK_IO_TIME    = "disk_io_time"
	SCHEDULING_TIME = "scheduling_time"
)

var polarisMetrics = map[string][]string{
	CPU_TIME:        {"cpu"},
	NETWORK_TIME:    {"net", "epoll"},
	LOCK_GC_TINME:   {"futex"},
	DISK_IO_TIME:    {"file"},
	SCHEDULING_TIME: {"runq"},
}

func CheckPolarisType(reason string) bool {
	if reason == CPU_TIME || reason == NETWORK_TIME || reason == LOCK_GC_TINME || reason == DISK_IO_TIME || reason == SCHEDULING_TIME {
		return true
	}
	return false
}

func GetPolarisMetrics(reason string) []string {
	return polarisMetrics[reason]
}
