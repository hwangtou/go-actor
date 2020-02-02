package go_actor

import (
	"errors"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	ErrMessageValue = errors.New("message value error")
)

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

func checkMessage(msg interface{}, export bool, lv int) error {
	if t := reflect.TypeOf(msg); t == nil {
		return ErrMessageValue
	} else {
		switch v := reflect.ValueOf(msg); t.Kind() {
		case reflect.Array, reflect.Map, reflect.Slice:
			if err := checkMessage(reflect.TypeOf(t).Elem(), export, lv+1); err != nil {
				return err
			}
		case reflect.Struct:
			if export {
				if rune, _ := utf8.DecodeRuneInString(t.Name()); !unicode.IsUpper(rune) {
					// struct is not exported
					return ErrMessageValue
				}
			}
			for i := 0; i < t.NumField(); i++ {
				if err := checkMessage(t.Field(i).Type, export, lv+1); err != nil {
					return err
				}
			}
		case reflect.Ptr, reflect.Interface:
			if lv > 0 {
				return ErrMessageValue
			}
			e := v.Elem()
			if e.Kind() == reflect.Invalid {
				return ErrMessageValue
			}
			if err := checkMessage(e.Interface(), export, lv); err != nil {
				return err
			}
		case reflect.Invalid, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return ErrMessageValue
		}
	}
	return nil
}

type messageType int

const (
	msgTypeSend   = 0
	msgTypeAsk    = 1
	msgTypeAnswer = 2
	msgTypeKill   = 3
)
