package twitchapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type files struct {
	Files []File
}

func findLatestMatchingFile(files []File, version string,
	releaseType int, modVersion int) (tfile File, seen bool) {
	var latestTime time.Time
	for _, file := range files {
		seenVersion := false
		for _, ver := range file.GameVersion {
			if ver == version {
				seenVersion = true
				break
			}
		}
		if seenVersion &&
			latestTime.Before(file.FileDate) &&
			(releaseType == -1 || releaseType >= file.ReleaseType) &&
			(modVersion == -1 || modVersion == file.ID) {
			seen = true
			tfile = file
			latestTime = file.FileDate
		}
	}
	return
}

// LatestDownload returns the latest download that fulfills certain conditions or an error if no such download exists.
// If releaseType or modVersion is -1, the respective condition is ignored.
func (info *ModInfo) LatestDownload(
	client *resty.Client, version string, releaseType int, modVersion int) (*File, error) {
	file, seen := findLatestMatchingFile(info.LatestFiles, version, releaseType, modVersion)
	if !seen {
		resp, err := client.R().
			SetResult(files{}.Files).
			Get(fmt.Sprintf("https://%v/api/v2/addon/%v/files", Endpoint, info.ID))
		if err != nil {
			return nil, errors.Wrap(err, "error fetching downloads for "+info.Name)
		}
		files := resp.Result().(*[]File)
		file, seen = findLatestMatchingFile(*files, version, releaseType, modVersion)
		if !seen {
			return nil, fmt.Errorf(`couldn't find a download for %v that satisfies:
Game Version: %v
Release Type: %v or lower
Mod Version ID: %v`, info.Name, version, releaseType, modVersion)
		}
	}
	return &file, nil
}

// ActualName returns the actual name of a mod, because CurseForge replaces spaces with pluses.
func (file *File) ActualName() string {
	return strings.ReplaceAll(file.FileName, " ", "+")
}
