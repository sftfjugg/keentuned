package modules

var detectSpecEnvDict = map[string]string{
	"innodb_log_file_size": "cpu_core",
}

var methodResultDict = map[string]map[string]string{
	"cpu_core": innodbLogFileSizeDict,
}

var innodbLogFileSizeDict = map[string]string{
	"2":       "1500M",
	"4":       "2048M",
	"8":       "4096M",
	"16":      "8192M",
	"32":      "10240M",
	"64":      "20480M",
	"default": "2048M",
}


