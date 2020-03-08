package supos_proto_collector_backend

const (
	ResultCode_Success     = iota
	ResultCode_Failed      = iota
	ResultCode_WaitingAuth = iota
	ResultCode_IllegalAuth
	ResultCode_ServerKickOff
	ResultCode_ServerUnexpect
	ResultCode_Confirmation
)
