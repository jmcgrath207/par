package storage

var recordMap = map[string]map[string]string{}

func SetRecord(recordType string, ipAddress string, data string) {
	recordMap[ipAddress][recordType] = data
}
func GetRecord(recordType string, ipAddress string) string {
	return recordMap[ipAddress][recordType]
}
