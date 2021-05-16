package fileshare

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/c2h5oh/datasize"
)

type UploadDescription struct {
	Name      string
	Extension string
	Size      int64
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
		"Name":      ud.Name,
		"Extension": ud.Extension,
		"Size":      strconv.FormatInt(ud.Size, 10),
		"FullName":  ud.GetFullName(),
		"URL":       ud.GetURL(),
		"HumanSize": ud.GetHumanSize(),
	}
	return json.Marshal(output)
}
