package spec

import (
	"path"

	"senan.xyz/g/gonic/model"
)

func NewAlbumByFolder(f *model.Album) *Album {
	return &Album{
		Artist:   f.Parent.RightPath,
		CoverID:  f.ID,
		ID:       f.ID,
		IsDir:    true,
		ParentID: f.ParentID,
		Title:    f.RightPath,
	}
}

func NewTCAlbumByFolder(f *model.Album) *TrackChild {
	trCh := &TrackChild{
		ID:       f.ID,
		IsDir:    true,
		Title:    f.RightPath,
		ParentID: f.ParentID,
	}
	if f.Cover != "" {
		trCh.CoverID = f.ID
	}
	return trCh
}

func NewTCTrackByFolder(t *model.Track, parent *model.Album) *TrackChild {
	trCh := &TrackChild{
		ID:          t.ID,
		Album:       t.Album.RightPath,
		ContentType: t.MIME(),
		Suffix:      t.Ext(),
		Size:        t.Size,
		Artist:      t.TagTrackArtist,
		Title:       t.TagTitle,
		TrackNumber: t.TagTrackNumber,
		DiscNumber:  t.TagDiscNumber,
		Path: path.Join(
			parent.LeftPath,
			parent.RightPath,
			t.Filename,
		),
		ParentID: parent.ID,
		Duration: t.Length,
		Bitrate:  t.Bitrate,
		IsDir:    false,
		Type:     "music",
	}
	if parent.Cover != "" {
		trCh.CoverID = parent.ID
	}
	return trCh
}

func NewArtistByFolder(f *model.Album) *Artist {
	return &Artist{
		ID:         f.ID,
		Name:       f.RightPath,
		AlbumCount: f.ChildCount,
	}
}

func NewDirectoryByFolder(f *model.Album, children []*TrackChild) *Directory {
	return &Directory{
		ID:       f.ID,
		Parent:   f.ParentID,
		Name:     f.RightPath,
		Children: children,
	}
}
