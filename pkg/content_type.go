package cloudsearch

import (
	"strings"
)

type ContentType string

const (
	Application ContentType = "Application"
	Calendar    ContentType = "Calendar"
	Contact     ContentType = "Contact"
	Document    ContentType = "Document"
	Email       ContentType = "Email"
	Event       ContentType = "Event"
	File        ContentType = "File"
	Folder      ContentType = "Folder"
	Image       ContentType = "Image"
	Message     ContentType = "Message"
	Post        ContentType = "Post"
	Task        ContentType = "Task"
	Video       ContentType = "Video"
)

var FileTypes = []ContentType{
	Image, Video, Folder, Document, File,
}

var ImageTypes = []string{
	"png", "jpg", "jpeg", "gif", "bmp", "svg",
}

var VideoTypes = []string{
	"mpg", "mkv", "avi", "mp4",
}

func ContainsType(contentTypes []ContentType, content ContentType) bool {
	for _, r := range contentTypes {
		if r == content {
			return true
		}
	}
	return false
}

func ContainsAnyType(contentTypes []ContentType, contents []ContentType) bool {
	for _, r := range contentTypes {
		for _, j := range contents {
			if r == j {
				return true
			}
		}
	}
	return false
}

func ContainFileTypes(contentTypes []ContentType) bool {
	if len(contentTypes) == 0 {
		return true
	}
	for _, f1 := range contentTypes {
		for _, f := range FileTypes {
			if f == f1 {
				return true
			}
		}
	}
	return false
}

// try mime, then extension, then the full path
func kindFor(isDir bool, mimeType string, extension string, path string) ContentType {
	if isDir {
		return Folder
	}
	if mimeType != "" {
		if strings.Contains(mimeType, "image") || strings.Contains(mimeType, "drawing") {
			return Image
		} else if strings.Contains(mimeType, "video") {
			return Video
		} else if strings.Contains(mimeType, "folder") {
			return Folder
		} else if strings.Contains(mimeType, "document") {
			return Document
		}
	}
	return contentKindForExtension(extension)
}

func contentKindForExtension(fileType string) ContentType {
	fileType = strings.ToLower(
		strings.TrimPrefix(fileType, "."),
	)

	if StringsContain(ImageTypes, fileType) {
		return Image
	}

	if StringsContain(VideoTypes, fileType) {
		return Video
	}

	return File
}
