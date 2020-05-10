package actor

import (
	"log"
	"testing"
)

var testSys *system

func TestRemote_DefaultInit(t *testing.T) {
	if err := Remote.Init(NodeConfig{
		Id: 1,
	}); err != nil {
		t.Error(err)
		return
	}
}

func TestRemote_SysInit(t *testing.T) {
	testSys = &system{}
	testSys.init()
	if err := testSys.Remote().Init(NodeConfig{
		Id:            2,
		ListenNetwork: TCP4,
		ListenAddress: "127.0.0.1:12346",
		Authorization: map[string]string{
			"": "",
		},
	}); err != nil {
		t.Error(err)
		return
	}
}

func TestRemote_DefaultConnSys(t *testing.T) {
	rConn, err := Remote.NewConn(2, "", "", TCP, "0.0.0.0:12346")
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
