package lib

import "time"

type JSONTime time.Time

func (jt JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(jt).Format(time.RFC3339)), nil
}

func (jt *JSONTime) UnmarshalJSON(bs []byte) error {
	parsed, err := time.Parse(time.RFC3339, string(bs))
	if err != nil {
		return err
	}

	*jt = JSONTime(parsed)
	return nil
}
