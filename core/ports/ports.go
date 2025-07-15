package ports

import(
	"github.com/Y-Figos/ipesdk/core/df"
)

type InputInterface interface{
	GetData() (*df.Dataframe, error) // The Module calls this
	Open() error
	ReadSample(n int) ([][]string, error) 
	Close() error
	GetHeaders() ([]string, error)
}
type InputInterfaceBatchReader interface {
	ReadBatch(n int) ([][]string, error) 
}
type OutputInterface interface{
	ExportData() error // The Module calls this
	// WriteHeaders() error
	// WriteData() error
	Open() error
	Close() error
}