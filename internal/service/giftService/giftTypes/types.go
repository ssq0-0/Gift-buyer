package giftTypes

type GiftResult struct {
	GiftID  int64
	Success bool
	Err     error
}

type GiftSummary struct {
	GiftID    int64
	Requested int64
	Success   int64
}
