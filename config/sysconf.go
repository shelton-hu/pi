package config

// SystemConfig ...
type SystemConfig struct {
	Registry   Registry          `json:"registry"`
	Mysql      map[string]Mysql  `json:"database"`
	Redis      Redis             `json:"redis"`
	Jaeger     Jaeger            `json:"jaeger"`
	Kafka      Kafka             `json:"kafka"`
	DelayQueue DelayQueue        `json:"delay_queue"`
	Cron       Cron              `json:"cron"`
	Domain     map[string]string `json:"domain"`
}

// Registry ...
type Registry struct {
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Ttl      int               `json:"ttl"`
	Interval int               `json:"interval"`
	Version  string            `json:"version"`
	MetaData map[string]string `json:"meta_data"`
}

// Mysql ...
type Mysql struct {
	Dialect        string `json:"dialect"`
	Database       string `json:"database"`
	User           string `json:"user"`
	Password       string `json:"password"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Charset        string `json:"charset"`
	MaxIdleConnNum int    `json:"max_idle_conn_num"`
	MaxOpenConnNum int    `json:"max_open_conn_num"`
}

// Redis ...
type Redis struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	MaxIdle     int    `json:"max_idle"`
	MaxActive   int    `json:"max_active"`
	IdleTimeout int    `json:"idle_timeout"`
	KeyPrefix   string `json:"key_prefix"`
}

// Jaeger ...
type Jaeger struct {
	Host string  `json:"host"`
	Port int     `json:"port"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

// Kafka ...
type Kafka struct {
	Addrs   []string `json:"addrs"`
	Version string   `json:"version"`
}

// DelayQueue ...
type DelayQueue struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Namespace string `json:"namespace"`
	Token     string `json:"token"`
}

// Cron ...
type Cron struct {
	Spec []CronSpec `json:"spec"`
}

// CronSpec is the spec of the cron.
type CronSpec struct {
	// Name of the task，need to be consistent with the name when registering the task.
	Name string `json:"name"`

	// Schedule is the schedule of the task，'@reboot' is not currently supported.
	// @wiki https://en.wikipedia.org/wiki/Cron
	Schedule string `json:"schedule"`

	// Suspend can be used to stop or start the task，when it's true, the task will
	// or has been stopped. Default value of the suspend is false.
	Suspend bool `json:"suspend"`

	// Parallelism is the concurrent number of the task.
	Parallelism int `json:"parallelism"`
}
