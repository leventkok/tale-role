package license

import "time"

type ProductLicense struct {
	ID        string
	UserID    string
	DeviceID  string
	Platform  string
	CreatedAt time.Time
}
