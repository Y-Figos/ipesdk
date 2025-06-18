package ports

import(
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
)

type InputInterface interface{
	GetData() (*df.Dataframe, error) // The Module calls this
	Open() error
	ReadSample(n int) ([][]string, error)
	ReadBatch(n int) ([][]string, error)  
	Close() error
	GetHeaders() ([]string, error)
}
type OutputInterface interface{
	ExportData() error // The Module calls this
	WriteHeaders() error
	WriteData() error
	Open() error
	Close() error
}