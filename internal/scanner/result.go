package scanner

// SourceResult captures the outcome of running a single passive source.
type SourceResult struct {
	Name    string
	Count   int
	Err     error
	Skipped bool // true if the source was never attempted (e.g. missing dependency)
}

// Summary aggregates the results of a full scan for final reporting.
type Summary struct {
	Domain           string
	Results          []SourceResult
	UniqueSubdomains int
	FinalFilePath    string
	Duration         string
}
