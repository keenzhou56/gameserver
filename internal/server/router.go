package server

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type HandlersChain struct {
	imType     uint16
	handleName string
	next       *HandlersChain
}

func (srv *Server) InitRouter() {
	srv.router = &HandlersChain{
		imType:     0,
		handleName: "",
	}
}

func (srv *Server) AddRouter(imType uint16, fn interface{}) {
	handleName := srv.nameOfFunction(fn)
	if srv.router.imType == 0 {
		srv.router.imType = imType
		srv.router.handleName = handleName
		return
	}
	newNode := srv.router.NewNode(imType, handleName)
	srv.router.insertNode(newNode)
}

func (srv *Server) FindRouter(imType uint16) string {
	handleName := ""
	temp := srv.router
	for {
		if temp.imType == imType {
			handleName = temp.handleName
			break
		}
		if temp.next == nil {
			break
		}
		temp = temp.next
	}
	return handleName
}

func (srv *Server) nameOfFunction(f interface{}) string {
	handlFuncFull := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	nameEnd := filepath.Ext(handlFuncFull)
	handlFunc := strings.TrimPrefix(nameEnd, ".")
	return strings.TrimSuffix(handlFunc, "-fm")
}

func (srv *Server) ListRouter() {
	temp := srv.router
	for {
		fmt.Println(temp.imType, temp.handleName)
		if temp.next == nil {
			break
		}
		temp = temp.next
	}
}

func (head *HandlersChain) NewNode(imType uint16, handleName string) *HandlersChain {
	newNode := &HandlersChain{
		imType:     imType,
		handleName: handleName,
	}
	return newNode
}

func (head *HandlersChain) insertNode(newNode *HandlersChain) {
	temp := head
	for {
		if temp.next == nil {
			break
		}
		temp = temp.next
	}
	temp.next = newNode
}

func (head *HandlersChain) sortInsertNode(newNode *HandlersChain) {
	temp := head
	flag := true
	for {
		if temp.next == nil {
			break
		} else if temp.next.imType > newNode.imType {
			break
		} else if temp.next.imType == newNode.imType {
			flag = false
			break
		}
		temp = temp.next
	}
	if flag {
		newNode.next = temp.next
		temp.next = newNode
	}
}

func (head *HandlersChain) delNode(imType uint16) {
	temp := head
	flag := false
	for {
		if temp.next == nil {
			break
		} else if temp.next.imType == imType {
			flag = true
			break
		}
		temp = temp.next
	}
	if flag {
		temp.next = temp.next.next
	}
}

func (head *HandlersChain) modifyNode(newNode *HandlersChain) {
	temp := head
	flag := false
	for {
		if temp.next == nil {
			break
		} else if temp.next.imType == newNode.imType {
			flag = true
			break
		}
		temp = temp.next
	}
	if flag {
		if temp.next.next != nil {
			newNode.next = temp.next.next
		}
		temp.next = newNode
	}
}
