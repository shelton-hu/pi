package config

type SystemConfig struct {
	Registry   Registry          `json:"registry"`
	Mysql      map[string]Mysql  `json:"database"`
	Redis      Redis             `json:"redis"`
	Jeager     Jeager            `json:"jeager"`
	Kafka      Kafka             `json:"kafka"`
	DelayQueue DelayQueue        `json:"delay_queue"`
	Domain     map[string]string `json:"domain"`
}

type Registry struct {
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Ttl      int               `json:"ttl"`
	Interval int               `json:"interval"`
	Version  string            `json:"version"`
	MetaData map[string]string `json:"meta_data"`
}

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

type Redis struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	MaxIdle     int    `json:"max_idle"`
	MaxActive   int    `json:"max_active"`
	IdleTimeout int    `json:"idle_timeout"`
	KeyPrefix   string `json:"key_prefix"`
}

type Jeager struct {
	Host string  `json:"host"`
	Port int     `json:"port"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

type Kafka struct {
	Addrs   []string `json:"addrs"`
	Version string   `json:"version"`
}

type DelayQueue struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Namespace string `json:"namespace"`
	Token     string `json:"token"`
}
