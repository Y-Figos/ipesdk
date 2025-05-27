package ports

import(
	"codehub-g.huawei.com/ProjectIPE/IPEGOCORE/core/df"
)

type InputInterface interface{
	GetData() (*df.Dataframe, error)
}

