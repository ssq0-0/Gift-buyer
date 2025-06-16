package giftBuyer

type giftResult struct {
	giftID  int64
	success bool
	err     error
}

type giftSummary struct {
	giftID    int64
	requested int64
	success   int64
	// errors    map[string]int64
}
