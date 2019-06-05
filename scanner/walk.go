package scanner

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/karrick/godirwalk"
	"github.com/pkg/errors"

	"github.com/sentriz/gonic/mime"
	"github.com/sentriz/gonic/model"
)

type item struct {
	fullPath  string
	relPath   string
	directory string
	filename  string
	stat      os.FileInfo
}

func (s *Scanner) callbackItem(fullPath string, info *godirwalk.Dirent) error {
	stat, err := os.Stat(fullPath)
	if err != nil {
		return errors.Wrap(err, "stating")
	}
	relPath, err := filepath.Rel(s.musicPath, fullPath)
	if err != nil {
		return errors.Wrap(err, "getting relative path")
	}
	directory, filename := path.Split(relPath)
	it := &item{
		fullPath:  fullPath,
		relPath:   relPath,
		directory: directory,
		filename:  filename,
		stat:      stat,
	}
	if info.IsDir() {
		return s.handleFolder(it)
	}
	lowerFilename := strings.ToLower(filename)
	if _, ok := coverFilenames[lowerFilename]; ok {
		s.curCover = filename
		return nil
	}
	ext := path.Ext(filename)[1:]
	if _, ok := mime.Types[ext]; ok {
		return s.handleTrack(it)
	}
	return nil
}

func (s *Scanner) callbackPost(fullPath string, info *godirwalk.Dirent) error {
	folder := s.curFolders.Pop()
	if folder.IsNew {
		folder.ParentID = s.curFolderID()
		folder.Cover = s.curCover
		s.tx.Save(folder)
	}
	s.curCover = ""
	log.Printf("processed folder `%s`\n", fullPath)
	return nil
}

func (s *Scanner) handleFolder(it *item) error {
	folder := &model.Album{}
	defer s.curFolders.Push(folder)
	err := s.tx.
		Where(model.Album{
			LeftPath:  it.directory,
			RightPath: it.filename,
		}).
		First(folder).
		Error
	if !gorm.IsRecordNotFoundError(err) &&
		it.stat.ModTime().Before(folder.UpdatedAt) {
		// we found the record but it hasn't changed
		return nil
	}
	folder.LeftPath = it.directory
	folder.RightPath = it.filename
	s.tx.Save(folder)
	folder.IsNew = true
	return nil
}

func (s *Scanner) handleTrack(it *item) error {
	//
	// set track basics
	track := &model.Track{}
	defer func() {
		// id will will be found (the first early return)
		// or created the tx.Save(track)
		s.seenTracks[track.ID] = struct{}{}
	}()
	err := s.tx.
		Where(model.Track{
			AlbumID:  s.curFolderID(),
			Filename: it.filename,
		}).
		First(track).
		Error
	if !gorm.IsRecordNotFoundError(err) &&
		it.stat.ModTime().Before(track.UpdatedAt) {
		// we found the record but it hasn't changed
		return nil
	}
	track.Filename = it.filename
	track.Size = int(it.stat.Size())
	track.AlbumID = s.curFolderID()
	track.Duration = -1
	track.Bitrate = -1
	tags, err := readTags(it.fullPath)
	if err != nil {
		return errors.Wrap(err, "reading tags")
	}
	trackNumber, totalTracks := tags.Track()
	discNumber, totalDiscs := tags.Disc()
	track.TagDiscNumber = discNumber
	track.TagTotalDiscs = totalDiscs
	track.TagTotalTracks = totalTracks
	track.TagTrackNumber = trackNumber
	track.TagTitle = tags.Title()
	track.TagTrackArtist = tags.Artist()
	track.TagYear = tags.Year()
	//
	// set album artist basics
	artist := &model.Artist{}
	err = s.tx.
		Where("name = ?", tags.AlbumArtist()).
		First(artist).
		Error
	if gorm.IsRecordNotFoundError(err) {
		artist.Name = tags.AlbumArtist()
		s.tx.Save(artist)
	}
	track.ArtistID = artist.ID
	s.tx.Save(track)
	//
	// set album if this is the first track in the folder
	if !s.curFolder().IsNew {
		return nil
	}
	s.curFolder().TagTitle = tags.Album()
	s.curFolder().TagYear = tags.Year()
	s.curFolder().TagArtistID = artist.ID
	return nil
}
