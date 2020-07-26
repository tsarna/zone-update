package updater

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/gofrs/flock"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zoneupdated/atomicfile"
	"zoneupdated/config"
	"zoneupdated/httperror"
)

type UpdateRequest struct {
	FQDN    string `json:"fqdn"`
	RRType  string `json:"rrtype"`
	Value   string `json:"value"`
	Disable bool   `json:"-"`
}

type Updater struct {
	conf          config.Config
	lockfile      *flock.Flock
	serialMatcher *regexp.Regexp
}

func New(conf config.Config) Updater {
	updater := Updater{
		conf:          conf,
		lockfile:      flock.New(fmt.Sprintf("%s.lock", conf.ZoneFileName)),
		serialMatcher: regexp.MustCompile("(?i)^(\\s*)(\\d+)(\\s*;\\s*serial\\s*)$"),
	}

	return updater
}

func (updater *Updater) Update(ctx context.Context, updateRequest UpdateRequest) error {
	success, err := updater.lockfile.TryLockContext(ctx, time.Second)
	if !success {
		if err == nil {
			return fmt.Errorf("unknown error")
		} else {
			return httperror.Error(http.StatusConflict, err)
		}
	}
	defer updater.lockfile.Unlock()

	zoneFile, err := os.Open(updater.conf.ZoneFileName)
	if err != nil {
		return fmt.Errorf("Unable to open zone file: %s", err)
	}
	defer zoneFile.Close()

	newZoneFile, err := atomicfile.Open(updater.conf.ZoneFileName)
	if err != nil {
		return fmt.Errorf("Unable to open temporary file: %s", err)
	}

	changed, err := updater.copyAndUpdate(zoneFile, newZoneFile, updateRequest)

	if updater.conf.TestMode {
		newZoneFile.Close()
	} else if changed {
		return newZoneFile.Commit()
	} else {
		_ = newZoneFile.Abort()
	}
	return err
}

func (updater *Updater) copyAndUpdate(currentFile io.Reader, newFile *atomicfile.AtomicFile, updateRequest UpdateRequest) (bool, error) {
	found := false
	changed := false

	hash := cNameHash(updateRequest.FQDN)
	newValue := updateRequest.Value
	numFields := len(strings.Fields(newValue))
	if numFields != 1 || strings.ContainsRune(newValue, '"') {
		// quote the string
		newValue = fmt.Sprint("\"", strings.ReplaceAll(newValue, "\"", "\\\""), "\"")
	}

	recordMatchRegex := fmt.Sprintf("(?i)^(\\s*;)?(\\s*)(%s|%s)(\\s*\\d+)?(\\s*IN)?(\\s*%s)(\\s*)",
		updateRequest.FQDN, hash, updateRequest.RRType)

	recordMatcher, err := regexp.Compile(recordMatchRegex)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(currentFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		// Update serial number
		groups := updater.serialMatcher.FindStringSubmatch(line)
		if groups != nil {
			serial, err := getSerial(groups[2])
			if err != nil {
				return false, err
			}

			if !updater.conf.SequentialSerial {
				timeSerial := timeBasedSerial()
				if timeSerial > serial {
					serial = timeSerial
				}
			}

			_, err = fmt.Fprintf(newFile, "%s%d%s\n", groups[1], serial+1, groups[3])
			if err != nil {
				return false, err
			}
		} else {
			var newLine string
			groups := recordMatcher.FindStringSubmatch(line)
			if groups != nil {
				found = true

				disableComment := ""
				if updateRequest.Disable {
					disableComment = ";"
				}

				newLine = fmt.Sprint(disableComment, groups[2], groups[3], groups[4], groups[5], groups[6], groups[7],
					newValue)

				if newLine != line {
					changed = true
				}
			} else {
				newLine = line
			}

			_, err = fmt.Fprintln(newFile, newLine)
			if err != nil {
				return false, err
			}
		}
	}

	if !found {
		msg := fmt.Sprintf("Did not find record for %s or %s with RRTYPE %s",
			updateRequest.FQDN, hash, updateRequest.RRType)
		log.Print(msg)
		return false, httperror.Error(http.StatusBadRequest,
			errors.New(msg))
	}

	return changed, nil
}

func cNameHash(fqdn string) string {
	sum := sha1.Sum([]byte(fqdn))
	return base32.StdEncoding.EncodeToString(sum[:])
}

func getSerial(serial string) (uint32, error) {
	stamp, err := strconv.ParseUint(serial, 10, 32)

	return uint32(stamp), err
}

func timeBasedSerial() uint32 {
	stamp, err := strconv.ParseUint(time.Now().Format("20060102"), 10, 32)
	if err != nil {
		return 0
	}

	return uint32(stamp * 100)
}
