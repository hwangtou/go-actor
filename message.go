package go_actor

//
// message
//

type message struct {
	sender     Ref
	msgSession uint64
	msgType    messageType
	msgContent interface{}
	msgError   error
}

type messageType int

const (
	msgTypeSend   = 0
	msgTypeAsk    = 1
	msgTypeAnswer = 2
	msgTypeKill   = 3
)
