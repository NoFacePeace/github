package server

import (
	"errors"
	"fmt"
)

const (
	connTypeMax = 8
)

var connTypes [connTypeMax]*connectionType

type connListener struct {
	ct    *connectionType
	count int
}

type connectionType struct {
	listen        func(listener *connListener) error
	getType       func(conn *connection) string
	init          func()
	acceptHandler aeFileProc
}

type connection struct{}

func connListen(listener *connListener) error {
	return listener.ct.listen(listener)
}

func connTypeInitialize() {
	redisRegisterConnectionTypeSocket()
}

func connTypeRegister(ct *connectionType) error {
	name := ct.getType(nil)
	var typ int
	for typ = 0; typ < connTypeMax; typ++ {
		if connTypes[typ] == nil {
			break
		}
		if name == connTypes[typ].getType(nil) {
			return fmt.Errorf("connection types %s already registered", name)
		}
	}
	if typ == connTypeMax {
		return errors.New("connetion type max limit")
	}
	connTypes[typ] = ct
	if ct.init != nil {
		ct.init()
	}
	return nil
}

func connAcceptHandler(ct *connectionType) aeFileProc {
	return ct.acceptHandler
}
