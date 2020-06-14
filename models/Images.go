package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// Image represent data stored about a speific image stored in loal storage
type Image struct {
	// The entity the image is associated with e.g. a User ID for an Avatar image
	EntityID int
	Filename string
	Filepath string
}

// Defines API for interacting with images
type ImageService interface {
	Create(id int, r io.Reader, filename string) error
	//Looks up Images associated with an EntityID
	GetByEntityID(entid int) ([]Image, error)
	Delete(i *Image) error
}

func NewImageService(path string) ImageService {
	return &imageModel{
		Filepath: path,
	}
}

type imageModel struct {
	Filepath string
}

// Path is used to build the absolute path used to reference this image
// via a web request.
func (i *Image) Path() string {
	temp := url.URL{Path: "/" + i.RelativePath()}
	return temp.String()
}

// RelativePath is used to build the path to this image on our local
// disk, relative to where our Go application is run from.
func (i *Image) RelativePath() string {
	// Convert the gallery ID to a string
	//id := fmt.Sprintf("%v", i.EntityID)
	return filepath.ToSlash(filepath.Join(i.Filepath, i.Filename))
}

func (is *imageModel) Create(id int, r io.Reader, filename string) error {
	path, err := is.mkImagePath(id)
	if err != nil {
		return err
	}
	// Create a destination file
	dst, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer dst.Close()
	// Copy reader data to the destination file
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (im *imageModel) GetByEntityID(id int) ([]Image, error) {
	path := im.imagePath(id)
	strings, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}
	// Setup the Image slice we are returning
	ret := make([]Image, len(strings))
	for i, imgStr := range strings {
		ret[i] = Image{
			Filename: filepath.Base(imgStr),
			EntityID: id,
			Filepath: filepath.Dir(imgStr),
		}
	}
	return ret, nil
}

func (im *imageModel) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}

// Going to need this when we know it is already made
func (im *imageModel) imagePath(id int) string {
	return filepath.Join("images", im.Filepath,
		fmt.Sprintf("%v", id))
}

// Use the imagePath method we just made
func (im *imageModel) mkImagePath(id int) (string, error) {
	path := im.imagePath(id)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}
	return path, nil
}
