package session

import (
	"reflect"
	"testing"
)

func TestSecureCookie_EncodeDecode(t *testing.T) {
	hashKey := []byte("1234567890")
	blockKey := []byte("0123456701234567" + "0123456701234567") // 2 * 16
	secureCookie, err := NewSecureCookie(hashKey, blockKey)
	if err != nil {
		t.Error(err)
	}

	data := make(map[string]interface{})
	data["user"] = "username"
	data["user_id"] = 1

	sessionName := "cookieName"

	encoded, err := secureCookie.Encode(sessionName, data)
	if err != nil {
		t.Error(err)
	}
	decoded := make(map[string]interface{})
	err = secureCookie.Decode(sessionName, encoded, &decoded)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(data, decoded) {
		t.Error("source message and decoded message are not equal")
	}

}
