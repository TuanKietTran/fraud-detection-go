package domain

// StructuringMinEvents is the named synthetic pattern: three unique events in
// the account window, each within the inclusive top 10% of the configured band.
const StructuringMinEvents = 3

func NearStructuringBand(amount, band int64) bool {
	return amount > 0 && band > 0 && amount <= band && amount >= band-band/10
}

// InferredStructuring takes a count of prior qualifying unique events; the
// current event is included only after classification. No float rounding occurs.
func InferredStructuring(amount, band int64, prior int) bool {
	return prior >= StructuringMinEvents-1 && NearStructuringBand(amount, band)
}
