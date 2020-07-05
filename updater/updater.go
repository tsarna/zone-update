package updater

import (
  "context"
  "crypto/sha1"
  "encoding/base32"
  "fmt"
  "github.com/gofrs/flock"
  "io"
  "net/http"
  "os"
  "time"
  "zoneupdated/atomicfile"
  "zoneupdated/config"
  "zoneupdated/httperror"
)

type UpdateRequest struct {
  FQDN     string `json:"fqdn"`
  RRType   string `json:"rrtype"`
  Value    string `json:"value"`
  Disable  bool   `json:"-"`
}

func Update(ctx context.Context, conf config.Config, updateRequest UpdateRequest) error {
  lockfile := flock.New(fmt.Sprintf("%s.lock", conf.ZoneFileName))

  success, err := lockfile.TryLockContext(ctx, time.Second)
  if !success {
    if err == nil {
      return fmt.Errorf("unknown error")
    } else {
      return httperror.Error(http.StatusConflict, err)
    }
  }
  defer lockfile.Unlock()

  zoneFile, err := os.Open(conf.ZoneFileName)
  if err != nil {
    return fmt.Errorf("Unable to open zone file: ", err)
  }
  defer zoneFile.Close()

  newZoneFile, err := atomicfile.Open(conf.ZoneFileName)
  if err != nil {
    return fmt.Errorf("Unable to open temporary file: ", err)
  }

  changed, err := copyAndUpdate(zoneFile, newZoneFile, updateRequest)
  if changed {
    return newZoneFile.Commit()
  }

  newZoneFile.Abort()
  return err
}

func copyAndUpdate(currentFile io.Reader, newFile *atomicfile.AtomicFile, updateRequest UpdateRequest) (bool, error) {
  return false, nil
}

func cNameHash(fqdn string) string {
  sum := sha1.Sum([]byte(fqdn))
  return base32.StdEncoding.EncodeToString(sum[:])
}
