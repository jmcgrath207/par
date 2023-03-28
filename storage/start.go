package storage

var RecordMap = map[string]map[string]string{}

func SetRecord(recordType string, hostname string, data string) {
	hostname = hostname + "."

	// Initialize the inner map if it does not exist.
	if RecordMap[hostname] == nil {
		RecordMap[hostname] = map[string]string{}
	}

	// Set the key-value pair in the inner map.
	RecordMap[hostname][recordType] = data
}

func GetRecord(recordType string, hostname string) (string, bool) {
	_, ok := RecordMap[hostname]
	if ok {
		val, ok := RecordMap[hostname][recordType]
		if ok {
			return val, true
		}

	}
	return "", false
}
