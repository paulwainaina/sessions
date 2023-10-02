package sessions

import (
	"testing"
)

var sessionManager = NewSessionsManager()

func Test_SessionManager(t *testing.T) {
	sessionManager.ManageSession()
}

func Test_Create(t *testing.T) {
	session := sessionManager.CreateNewSession("paulwainaina@gmail.com")
	if session == nil {
		t.Fatal("Session not created")
	}
	f, session := sessionManager.GetSession("paulwainaina@gmail.com", session.Cookie.Value)
	if !f {
		t.Fatalf("Session was not found got ")
	}
	f, session = sessionManager.GetSession("paulwainaina@gmail.com1", session.Cookie.Value)
	if f || session != nil {
		t.Fatalf("Expected no session")
	}
	f, session = sessionManager.GetSession("paulwainaina@gmail.com", "HFJFJNKNDFTYFEWKFJFOIEWU")
	if f || session != nil {
		t.Fatalf("Expected no session")
	}
}
func Test_Update(t *testing.T) {
	updated, _ := sessionManager.UpdateSession("paulwainaina@gmail.com")
	if !updated {
		t.Fatal("expected a new token")
	}
	updated, _ = sessionManager.UpdateSession("paulwainaina@gmail.com")
	if updated {
		t.Fatal("no new token was expected")
	}
}

func Test_Delete(t *testing.T) {
	deleted := sessionManager.DeleteSession("paulwainaina@gmail.com")
	if !deleted {
		t.Fatal("expected to delete session ")
	}
}
