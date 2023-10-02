### Session Management

This is a go module meant to handle user tokens meant to be set on browsers.

1. The Module creates a session and stores it in a file for persistence. The path can be set based on the developers liking. The defaultis sessions.txt

    ```go
        import "github.com/paulwainaina/session"

        func GenerateCookie() http.Cookie {
            return http.Cookie{
                Name:     "session",
                Value:    uuid.NewString(),
                Path:     "/",
                Expires:  time.Now().Add(time.Duration(1) * time.Minute),
                Secure:   true,
                SameSite: http.SameSiteNoneMode,
            }
        }
    ```

2. The module supports the CRUD operations on the file.
    1. DeleteSession
    2. UpdateSession
    3. WriteSession
    4. SearchSession

3. There are mutex to lock the file while more than two go routines try to perform update,create and delete operations.

     ```go
        var lockfile sync.Mutex

        func DeleteSession(done chan bool, user string) {
            lockfile.Lock()
            defer lockfile.Unlock()
            file, err := os.OpenFile(fileName, os.O_RDWR, 0600)
   ```

4. The module also trys to improve on performance by truncating data from an offset based on the affected areas of the file.This will reduce the number of writes significantly.
   ```go 
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
   ```
