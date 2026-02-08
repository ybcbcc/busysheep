package service

import "time"

var cnLoc *time.Location

func getCNLoc() *time.Location {
	if cnLoc != nil {
		return cnLoc
	}
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		cnLoc = loc
	} else {
		cnLoc = time.FixedZone("CST", 8*3600)
	}
	return cnLoc
}

func nowCN() time.Time {
	return time.Now().In(getCNLoc())
}
