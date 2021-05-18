package fileshare

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/c2h5oh/datasize"
)

type UploadDescription struct {
	Name      string
	Extension string
	Size      int64
	Expiry    time.Time
}

func (ud *UploadDescription) GetFullName() string {
	return ud.Name + ud.Extension
}

func (ud *UploadDescription) GetURL() string {
	return fmt.Sprintf("/raw/%s", ud.GetFullName())
}

func (ud *UploadDescription) GetHumanSize() string {
	return datasize.ByteSize(ud.Size).HumanReadable()
}

func (ud *UploadDescription) GetJSON() ([]byte, error) {
	output := map[string]string{
		"FullName":  ud.GetFullName(),
		"URL":       ud.GetURL(),
		"HumanSize": ud.GetHumanSize(),
		"Expiry":    ud.Expiry.Format("2006-01-02 15:04:05"),
	}
	return json.Marshal(output)
}
