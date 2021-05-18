package fileshare

import (
	"encoding/json"
	"fmt"
	"github.com/hako/durafmt"
	"log"
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

func (ud *UploadDescription) GetHumanExpiry() string {
	baseTime := time.Time{}
	if ud.Expiry == baseTime {
		return "No Expiry"
	}
	diff := ud.Expiry.Sub(time.Now())
	log.Printf("%v", diff.Truncate(time.Hour).Hours())
	if diff.Truncate(time.Hour).Hours() > 0 {
		log.Printf("Hours > 0: %s", diff)
		return durafmt.Parse(diff).LimitFirstN(2).String()
	} else if diff.Truncate(time.Minute).Minutes() > 1 {
		log.Printf("Minutes > 1: %s", diff)
		return durafmt.Parse(diff).LimitFirstN(1).String()
	} else {
		log.Printf("Else: %s", diff)
		return durafmt.Parse(diff).LimitFirstN(1).String()
	}
}

func (ud *UploadDescription) GetJSON() ([]byte, error) {
	output := map[string]string{
		"FullName":  ud.GetFullName(),
		"URL":       ud.GetURL(),
		"HumanSize": ud.GetHumanSize(),
		"Expiry":    ud.Expiry.Format("2006-01-02 15:04:05"),
		"HumanExpiry": ud.GetHumanExpiry(),

	}
	return json.Marshal(output)
}
