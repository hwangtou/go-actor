package actor

import (
	"log"
	"testing"
)

var testSys *system
var testAllowedNames map[string]string = map[string]string{
	"default": "",
}

func TestRemote_DefaultInit(t *testing.T) {
	if err := Global.DefaultInit(1); err != nil {
		t.Error(err)
		return
	}
}

func TestRemote_SysInit(t *testing.T) {
	testSys = &system{}
	testSys.init()
	if err := testSys.Global().Init(2, TCP, "0.0.0.0:12346", testAllowedNames); err != nil {
		t.Error(err)
		return
	}
}

func TestRemote_DefaultConnSys(t *testing.T) {
	rConn, err := Global.NewConn(2, "default", "", TCP, "0.0.0.0:12346")
	if err != nil {
		t.Error(err)
		return
	}
	rRef, err := rConn.ByName("test")
	if err != nil {
		t.Error(err)
		return
	}
	log.Println(rRef)
}
