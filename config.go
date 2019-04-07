package status

import "github.com/cevaris/status/report/fileio"

var ApiTestStore = map[string]int64{
	"fileio_write_text": 0,
}

var Lookup = map[string]func(){
	"fileio_write_text": fileio.WriteTextReport,
}
