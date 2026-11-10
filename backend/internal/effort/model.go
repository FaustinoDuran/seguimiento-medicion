package effort

import (
	"strconv"
	"time"
)

// Hours stores an effort duration in hundredths of an hour.
type Hours int64

func (h Hours) String() string {
	value := int64(h)
	if value < 0 {
		return "-" + Hours(-value).String()
	}
	whole := value / 100
	frac := value % 100
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	if frac%10 == 0 {
		return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac/10, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac+100, 10)[1:]
}

type Entry struct {
	ID        int64
	ProjectID int64
	StoryID   int64
	MemberID  int64
	Date      time.Time
	Activity  string
	Hours     Hours
	CreatedAt time.Time
}

type RegisterInput struct {
	ProjectID int64
	StoryID   int64
	MemberID  int64
	Date      time.Time
	Activity  string
	Hours     string
}
