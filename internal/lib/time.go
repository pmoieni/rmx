package lib

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Timestamp time.Time

func GetTimestamp() Timestamp {
	return Timestamp(time.Now().UTC())
}

func (t Timestamp) ToSTDTime() time.Time {
	return time.Time(t)
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

func (t *Timestamp) UnmarshalJSON(bs []byte) error {
	parsed, err := time.Parse(time.RFC3339, string(bs))
	if err != nil {
		return err
	}

	*t = Timestamp(parsed)
	return nil
}

// Scan implements the Scanner interface for Timestamp
func (t *Timestamp) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*t = Timestamp(v)
		return nil
	case []byte:
		return t.scanString(string(v))
	case string:
		return t.scanString(v)
	default:
		return fmt.Errorf("cannot scan type %T into Timestamp", value)
	}
}

// scanString is a helper function to parse string timestamps
func (t *Timestamp) scanString(value string) error {
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	*t = Timestamp(parsedTime)
	return nil
}

// Value implements the driver Valuer interface for Timestamp
func (t Timestamp) Value() (driver.Value, error) {
	return time.Time(t).Format(time.RFC3339), nil
}

// String returns the time in RFC3339 format
func (t Timestamp) String() string {
	return time.Time(t).Format(time.RFC3339)
}
