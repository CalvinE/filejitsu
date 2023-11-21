package spaceanalyzer

type ScanParams struct {
	RootPath            string
	MaxRecursion        int
	CalculateFileHashes bool
	ConcurrencyLimit    int
}
