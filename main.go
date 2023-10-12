package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	fileName    = "./sessions.txt"
	lockfile    sync.Mutex
	sessionTime = 30
)

func init() {
	x, err := strconv.Atoi(os.Getenv("Session_Time"))
	if err == nil {
		sessionTime = x
	}
	var path = os.Getenv("Session_File")
	if len(path) != 0 {
		fileName = path
	}
}

type Session struct {
	User   string
	Cookie http.Cookie
}

func GenerateCookie() http.Cookie {
	return http.Cookie{
		Name:     "session",
		Value:    uuid.NewString(),
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(sessionTime) * time.Minute),
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		MaxAge:   sessionTime * 3600,
	}
}

func NewSession(user string) *Session {
	cookie := GenerateCookie()
	return &Session{User: user, Cookie: cookie}
}

type SessionsManager struct {
}

func SearchSession(readChannel chan *Session, user string) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		readChannel <- nil
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var session Session
		var text = scanner.Text()
		if strings.EqualFold(text, "") {
			continue
		}
		err = json.Unmarshal([]byte(text), &session)
		if err != nil {
			continue
		}
		if strings.EqualFold(session.User, user) {
			readChannel <- &session
			return
		}
	}
	readChannel <- nil
}

func DeleteSession(done chan bool, user string) {
	lockfile.Lock()
	defer lockfile.Unlock()
	file, err := os.OpenFile(fileName, os.O_RDWR, 0600)
	if err != nil {
		done <- false
	}
	defer file.Close()
	var data = []string{}
	scanner := bufio.NewScanner(file)
	var offset int64 = 0
	var found bool = false
	var numberofLines = 0
	var foundLine = 0
	for scanner.Scan() {
		numberofLines += 1
		var session Session
		var text = scanner.Text()
		if strings.EqualFold(text, "") {
			continue
		}
		err = json.Unmarshal([]byte(text), &session)
		if err != nil {
			fmt.Println(text)
			offset += int64(len([]byte(text)))
			continue
		}
		if !strings.EqualFold(session.User, user) {
			if !found {
				offset += int64(len([]byte(text)))
			} else {
				data = append(data, text)
			}
		} else {
			found = true
			foundLine = numberofLines
		}
	}
	if found {
		writter := bufio.NewWriter(file)
		if foundLine != 1 {
			writter.WriteString("\n")
			offset++
		}
		for _, x := range data {
			writter.WriteString(x + "\n")
		}
		file.Truncate(offset)
		file.Seek(offset, 0)
		err = writter.Flush()
		if err != nil {
			done <- false
			return
		}
	}
	done <- true
}

func UpdateSession(readChannel chan *Session, user string) {
	lockfile.Lock()
	defer lockfile.Unlock()
	file, err := os.OpenFile(fileName, os.O_RDWR, 0600)
	if err != nil {
		readChannel <- nil
		return
	}
	defer file.Close()
	var data = []string{}
	var newSession *Session = nil
	scanner := bufio.NewScanner(file)
	var offset int64 = 0
	var found bool = false
	var numberofLines = 0
	var foundLine = 0
	for scanner.Scan() {
		numberofLines += 1
		var session Session
		var text = scanner.Text()
		if strings.EqualFold(text, "") {
			continue
		}
		err = json.Unmarshal([]byte(text), &session)
		if err != nil {
			offset += int64(len([]byte(text)))
			continue
		}
		if !strings.EqualFold(session.User, user) {
			if !found {
				offset += int64(len([]byte(text)))
			} else {
				data = append(data, text)
			}
		} else {
			foundLine = numberofLines
			found = true
			session.Cookie = GenerateCookie()
			newSession = &session
			res, err := json.Marshal(session)
			if err != nil {
				continue
			}
			data = append(data, string(res))
		}
	}
	if found {
		writter := bufio.NewWriter(file)
		if foundLine != 1 {
			writter.WriteString("\n")
			offset++
		}
		for _, x := range data {
			writter.WriteString(x + "\n")
		}
		file.Truncate(offset)
		file.Seek(offset, 0)
		writter.Flush()
	}
	readChannel <- newSession
}

func WriteSession(writeChannel chan *Session, user string) {
	session := make(chan *Session)
	go SearchSession(session, user)
	result := <-session
	if result != nil {
		writeChannel <- result
		return
	}
	lockfile.Lock()
	defer lockfile.Unlock()
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		writeChannel <- nil
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	newSession := Session{User: user, Cookie: GenerateCookie()}
	jsonData, err := json.Marshal(newSession)
	if err != nil {
		writeChannel <- nil
		return
	}
	_, err = writer.WriteString(string(jsonData) + "\n")
	if err != nil {
		writeChannel <- nil
		return
	}
	writer.Flush()
	writeChannel <- &newSession
}

func (sessionsManager *SessionsManager) CreateNewSession(user string) *Session {
	/*
		for _, sess := range sessionsManager.Sessions {
			if strings.EqualFold(sess.User, user) {
				return sess
			}
		}
		session := NewSession(user)
		sessionsManager.Sessions = append(sessionsManager.Sessions, session)
		return session
	*/
	session := make(chan *Session)
	go WriteSession(session, user)
	return <-session
}

func (sessionsManager *SessionsManager) DeleteSession(user string) bool {
	/*for i, sess := range sessionsManager.Sessions {
		if strings.EqualFold(sess.User, user) {
			sessionsManager.Sessions = append(sessionsManager.Sessions[:i], sessionsManager.Sessions[i+1:]...)
			return
		}
	}*/
	deleted := make(chan bool)
	go DeleteSession(deleted, user)
	return <-deleted
}

func (sessionsManager *SessionsManager) GetSession(user, value string) (bool, *Session) {
	/*
		for _, sess := range sessionsManager.Sessions {
			if strings.EqualFold(sess.User, user) {
				if strings.EqualFold(sess.Cookie.Value, value) {
					return true, sess
				}
				break
			}
		}
		return false, nil
	*/
	session := make(chan *Session)
	go SearchSession(session, user)
	result := <-session
	if result == nil || !strings.EqualFold(result.Cookie.Value, value) {
		return false, nil
	}
	return true, result
}

func (sessionsManager *SessionsManager) UpdateSession(user string) (bool, *Session) {
	/*for _, sess := range sessionsManager.Sessions {
		if strings.EqualFold(sess.User, user) {
			sess.Cookie = GenerateCookie()
			return true, sess
		}
	}
	return false, nil*/
	session := make(chan *Session)
	go UpdateSession(session, user)
	result := <-session
	if result == nil {
		return false, nil
	}
	return true, result
}

func (sessionsManager *SessionsManager) ManageSession() {
	/*for {
		for i, sess := range sessionsManager.Sessions {
			if sess.Cookie.Expires.Before(time.Now()) {
				sessionsManager.Sessions = append(sessionsManager.Sessions[:i], sessionsManager.Sessions[i+1:]...)
				i--
			}
		}
	}*/
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	defer file.Close()
	for {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var session Session
			var text = scanner.Text()
			err = json.Unmarshal([]byte(text), &session)
			if err != nil {
				continue
			}
			if session.Cookie.Expires.Before(time.Now()) {
				done := make(chan bool)
				go DeleteSession(done, session.User)
				<-done
			}
		}
	}
}

func NewSessionsManager() *SessionsManager {
	return &SessionsManager{}
}
