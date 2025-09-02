package main

// tgConfig определяет структуру конфигурации для Telegram бота.
type tgConfig struct {
	Debug          bool   // Debug определяет, включен ли отладочный режим.
	Token          string // Token представляет токен Telegram бота.
	PoolAddress    string // PoolAddress представляет адрес пула Telegram.
	PoolPort       uint16 // PoolPort представляет порт пула Telegram.
	HookDomain     string // HookDomain представляет домен веб хука для Telegram.
	HookPort       uint16 // HookPort представляет порт вебхука для Telegram.
	HookCertPub    string // HookCertPub представляет путь к публичному ключу сертификата для вебхука.
	HookCertKey    string // HookCertKey представляет путь к закрытому ключу сертификата для вебхука.
	NameBot        string `yaml:"-"` // NameBot представляет имя бота (игнорируется при сериализации в YAML).
	WebHookAddress string `yaml:"-"` // WebHookAddress представляет адрес вебхука (игнорируется при сериализации в YAML).
}

// ldapConfig определяет структуру конфигурации для LDAP.
type ldapConfig struct {
	User     string   // User представляет имя пользователя для подключения к LDAP.
	Password string   // Password представляет пароль пользователя для подключения к LDAP.
	Servers  []string // Servers представляет список серверов LDAP.
	Dn       string   // Dn представляет базовый DN (Distinguished Name) для запросов LDAP.
}

// ldapUser определяет структуру данных для пользователя LDAP.
type ldapUser struct {
	TelegramId     int64  // TelegramId представляет идентификатор Telegram пользователя.
	DisplayName    string // DisplayName представляет отображаемое имя пользователя.
	SAMAccountName string // SAMAccountName представляет имя учетной записи пользователя в Active Directory.
}

// radiusConfig определяет структуру конфигурации для сервера Radius.
type radiusConfig struct {
	Debug         bool   // Debug определяет, включен ли отладочный режим.
	Address       string // Address представляет IP-адрес сервера Radius.
	Port          uint16 // Port представляет порт сервера Radius.
	Secret        string // Secret представляет общий секрет для авторизации на сервере Radius.
	AnswerTimeout int    // AnswerTimeout представляет время ожидания ответа от сервера Radius.
	ServerAddress string `yaml:"-"` // ServerAddress представляет адрес сервера Radius (игнорируется при сериализации в YAML).
}

// Cache определяет структуру для настройки кэша.
type cache struct {
	Timeout int // Timeout представляет время жизни записей в кэше.
}

// ConfigS определяет структуру для общей конфигурации приложения.
type configS struct {
	Telegram tgConfig     `yaml:"telegram"` // Telegram представляет конфигурацию Telegram.
	Ldap     ldapConfig   `yaml:"ldap"`     // Ldap представляет конфигурацию LDAP.
	Radius   radiusConfig `yaml:"radius"`   // Radius представляет конфигурацию сервера Radius.
	Cache    cache        `yaml:"cache"`    // Cache представляет конфигурацию кэша.
}
