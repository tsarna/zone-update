package restapi

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

type User struct {
	Username string
	Password string
}

type userMap map[string]User

type PasswordFile struct {
	filename    string
	credentials atomic.Value
}

func (p *PasswordFile) CheckPassword(_ context.Context, username string, password string) (bool, context.Context) {
	users := p.credentials.Load().(userMap)
	user, ok := users[username]
	if ok {
		if password == user.Password {
			return true, nil
		}
	}
	return false, nil
}

func NewPasswordFile(filename string) (*PasswordFile, error) {
	p := PasswordFile{filename: filename}
	err := p.loadFile(filename)
	return &p, err
}

func (p *PasswordFile) Reload() error {
	return p.loadFile(p.filename)
}

func (p *PasswordFile) loadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	lineNumber := 1

	credentials := make(userMap)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		} else if len(fields) == 2 {
			user := User{Username: fields[0], Password: fields[1]}
			credentials[user.Username] = user
		} else {
			return fmt.Errorf("entry needs two fields at line %d: '%s'", lineNumber, line)
		}

		lineNumber++
	}

	p.credentials.Store(credentials)

	return nil
}
