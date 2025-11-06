package utils

type CmdTuple struct {
	Flag string
	Data string
}

func NewCmdTuple(flag, data string) *CmdTuple {
	return &CmdTuple{
		Flag: flag,
		Data: data,
	}
}