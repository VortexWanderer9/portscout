package scanner

var commonServices = map[int]string{
	21:    "ftp",
	22:    "ssh",
	23:    "telnet",
	25:    "smtp",
	53:    "dns",
	69:    "tftp",
	80:    "http",
	123:   "ntp",
	110:   "pop3",
	143:   "imap",
	161:   "snmp",
	389:   "ldap",
	443:   "https",
	465:   "smtps",
	587:   "submission",
	636:   "ldaps",
	993:   "imaps",
	995:   "pop3s",
	1433:  "mssql",
	1521:  "oracle",
	2375:  "docker",
	3000:  "http-dev",
	3306:  "mysql",
	3389:  "rdp",
	5432:  "postgresql",
	5672:  "amqp",
	5900:  "vnc",
	6443:  "kubernetes",
	6379:  "redis",
	8080:  "http-alt",
	8443:  "https-alt",
	9200:  "elasticsearch",
	11211: "memcached",
	27017: "mongodb",
}

// ServiceName returns a well-known service name for a port, or "" if unknown.
func ServiceName(port int) string {
	return commonServices[port]
}
