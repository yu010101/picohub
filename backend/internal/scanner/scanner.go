package scanner

// Scanner defines the interface for malware scanning.
type Scanner interface {
	Scan(filePath string) (ScanResult, error)
}

type ScanResult struct {
	Clean   bool   `json:"clean"`
	Message string `json:"message"`
}

// NoopScanner always returns clean results (for development).
type NoopScanner struct{}

func NewNoopScanner() *NoopScanner {
	return &NoopScanner{}
}

func (s *NoopScanner) Scan(filePath string) (ScanResult, error) {
	return ScanResult{Clean: true, Message: "scan skipped (noop scanner)"}, nil
}
