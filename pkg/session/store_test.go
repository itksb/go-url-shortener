package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieStore(t *testing.T) {
	hashKey := []byte("1234567890")
	blockKey := []byte("0123456701234567" + "0123456701234567") // 2 * 16
	codec, err := NewSecureCookie(hashKey, blockKey)
	if err != nil {
		t.Fatal(err)
	}

	store := NewCookieStore(codec)
	request := httptest.NewRequest(
		"GET",
		"http://some.url",
		nil,
	)

	sessionName := "session"
	session, err := store.Get(request, sessionName)
	if err != nil {
		t.Fatal(err)
	}

	session.Values["user"] = "User"

	writer := httptest.NewRecorder()
	err = session.Save(request, writer)
	if err != nil {
		t.Fatal(err)
	}

	res := writer.Result()
	defer res.Body.Close()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected only one cookie, got: %d cookies instead.", len(cookies))
	}

	request2 := httptest.NewRequest(
		"GET",
		"http://some.url",
		nil,
	)
	request2.AddCookie(&http.Cookie{
		Name:  sessionName,
		Value: cookies[0].Value,
	})

	session, err = store.Get(request2, sessionName)
	if err != nil {
		t.Fatal(err)
	}

	valueFromSessionWithoutType := session.Values["user"]
	t.Logf("value from the session without type assertion: %v", valueFromSessionWithoutType)
	valueFromSession, ok := valueFromSessionWithoutType.(string)
	if !ok {
		t.Fatal("Type assertion to string failed or session value could not be retrieved.")
	}

	if valueFromSession != "User" {
		t.Fatalf("values extracted from cookie are not equal. \nExpected: %s \nGot: %s", "User", valueFromSession)
	}

}
