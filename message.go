package go_actor

//
// message
//

type message struct {
	sender     Ref
	sequence   uint64
	msgType    int
	msgContent interface{}
	msgError   error
}

const (
	msgTypeSend = 0
	msgTypeAsk = 1
)
