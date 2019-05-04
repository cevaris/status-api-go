package fileio

// ResponseJSON is the common JSON response for https://file.io
// {"success":true,"key":"tt67yI","link":"https://file.io/tt67yI","expiry":"14 days"}
// {"success":false,"error":404,"message":"Not Found"}
type ResponseJSON struct {
	Success bool `json:"success"`
}
