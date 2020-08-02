package restapi_test

import (
	"context"
	"fmt"
	"github.com/tsarna/zone-update/restapi"
	"io"
	"os"
	"testing"
)

func TestPasswordFile_CheckPassword(t *testing.T) {
	p, err := restapi.NewPasswordFile("testdata/test_passwd")
	if err != nil {
		t.Fatalf("Could not open password file: %s", err)
	}

	tryLogin(t, p, "user1", "password1", true)
	tryLogin(t, p, "user2", "password2", true)
	tryLogin(t, p, "user1", "password2", false)
	tryLogin(t, p, "user2", "password1", false)
	tryLogin(t, p, "user3", "password3", false)
	tryLogin(t, p, "user3", "password2", false)
	tryLogin(t, p, "user1", "password3", false)
}

func TestPasswordFile_DoesntExist(t *testing.T) {
	_, err := restapi.NewPasswordFile("testdata/NoSuchFile.oops")
	if err == nil {
		t.Error("Should have failed to open nonexistent password file")
	}
}

func TestPasswordFile_Reload(t *testing.T) {
	tempFile := fmt.Sprintf("%s%c%d.passwd", os.TempDir(), os.PathSeparator, os.Getpid())
	err := copyFile("testdata/test_passwd", tempFile)
	if err != nil {
		t.Fatalf("Unable to create temporary password file %s: %s", tempFile, err)
	}
	defer os.Remove(tempFile)

	p, err := restapi.NewPasswordFile(tempFile)
	if err != nil {
		t.Fatalf("Could not open password file: %s", err)
	}

	tryLogin(t, p, "user1", "password1", true)
	tryLogin(t, p, "user3", "password3", false)

	err = copyFile("testdata/invalid_passwd", tempFile)
	if err != nil {
		t.Fatalf("Unable to create replace password file %s: %s", tempFile, err)
	}

	err = p.Reload()
	if err == nil {
		t.Error("Reloading should have failed due to invalid file contents")
	}

	// Effective file password file contents should be unaffected due to failed reload
	tryLogin(t, p, "user1", "password1", true)
	tryLogin(t, p, "user3", "password3", false)

	err = copyFile("testdata/new_passwd", tempFile)
	if err != nil {
		t.Fatalf("Unable to create replace password file %s: %s", tempFile, err)
	}

	err = p.Reload()
	if err != nil {
		t.Errorf("Reloading new valid file should have worked but got: %s", err)
	}

	// Now the new passwords should be in effect
	tryLogin(t, p, "user1", "password1", false)
	tryLogin(t, p, "user3", "password3", true)
}

func tryLogin(t *testing.T, p *restapi.PasswordFile, user string, password string, expectOk bool) {
	ctx := context.TODO()

	ok, _ := p.CheckPassword(ctx, user, password)
	if ok != expectOk {
		t.Errorf("Login for user '%s' with password '%s' expected login success %t but got %t",
			user, password, expectOk, ok)
	}
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
