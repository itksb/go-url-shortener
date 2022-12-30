package filestorage

import (
	"bufio"
	"context"
	"fmt"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type storage struct {
	logger    logger.Interface
	fileRead  *os.File
	fileWrite *os.File
	reader    *bufio.Scanner

	currentURLID int64
	mtx          sync.RWMutex
}

// NewStorage - constructor
func NewStorage(logger logger.Interface, filename string) (*storage, error) {
	fileRead, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	lastID, err := getLastIDOrDefault(fileRead)
	if err != nil {
		logger.Error(fmt.Sprintf("filestorage:getLastIDOrDefault error: %s", err.Error()))
		return nil, err
	}

	logger.Info(fmt.Sprintf("last ID: %d", lastID))

	fileWrite, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0777)
	if err != nil {
		logger.Error(fmt.Sprintf("filestorage: open fileWrite error: %s", err.Error()))
		return nil, err
	}

	s := &storage{
		logger:       logger,
		fileWrite:    fileWrite,
		fileRead:     fileRead,
		reader:       bufio.NewScanner(fileRead),
		currentURLID: lastID,
		mtx:          sync.RWMutex{},
	}

	return s, nil
}

func (s *storage) Close() error {
	err1 := s.fileRead.Close()
	err2 := s.fileWrite.Close()
	if err1 == nil && err2 == nil {
		return nil
	}
	return fmt.Errorf("fileRead: %s. fileWrite: %s", err1.Error(), err2.Error())
}

func (s *storage) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.currentURLID++
	id := s.currentURLID

	if _, ok := s.findByID(id); ok {
		return "", fmt.Errorf("url with id %d already exists", id)
	}

	if err := s.persist(id, url, userID); err != nil {
		s.logger.Error(err.Error())
		return "", err
	}

	return strconv.FormatInt(id, 10), nil
}

func (s *storage) GetURL(ctx context.Context, id string) (string, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", err
	}

	if url, ok := s.findByID(idInt64); ok {
		return url, nil
	}

	return "", nil
}

func (s *storage) ListURLByUserID(ctx context.Context, userID string) ([]shortener.URLListItem, error) {
	var line string
	var foundItems []shortener.URLListItem
	_, err := s.fileRead.Seek(0, io.SeekStart)
	if err != nil {
		s.logger.Error(fmt.Sprintf("filestorage: fileWrite.Seek error. Err: %s", err.Error()))
		return foundItems, err
	}

	reader := bufio.NewScanner(s.fileRead)
	for reader.Scan() {
		line = reader.Text()
		curID, url, user, ok := extractValuesFromTheLine(line)
		if ok && user == userID {
			foundItems = append(foundItems, shortener.URLListItem{
				ID:          curID,
				UserID:      user,
				ShortURL:    "",
				OriginalURL: url,
			})
		}
	}

	err = s.reader.Err()
	if err != nil {
		s.logger.Error(fmt.Sprintf("filestorage: reader.Scan() error. Err: %s", err.Error()))
		return foundItems, err
	}

	return foundItems, nil
}

func (s *storage) findByID(id int64) (string, bool) {
	var foundValue, line string
	_, err := s.fileRead.Seek(0, io.SeekStart)
	if err != nil {
		s.logger.Error(fmt.Sprintf("filestorage: fileWrite.Seek error. Err: %s", err.Error()))
	}

	reader := bufio.NewScanner(s.fileRead)
	for reader.Scan() && len(foundValue) == 0 {
		line = reader.Text()
		curID, url, _, ok := extractValuesFromTheLine(line)
		if ok && curID == id {
			foundValue = url
		}
	}

	err = s.reader.Err()
	if err != nil {
		s.logger.Error(fmt.Sprintf("filestorage: reader.Scan() error. Err: %s", err.Error()))
	}

	return foundValue, len(foundValue) != 0 && err == nil
}

func extractValuesFromTheLine(line string) (int64, string, string, bool) {
	res := strings.SplitN(line, "::", 3)
	if len(res) == 3 {
		curID, err := strconv.ParseInt(res[0], 10, 64)
		if err == nil {
			return curID, res[1], res[2], true
		}
	}
	return 0, "", "", false
}

func (s *storage) persist(id int64, value string, userID string) error {
	_, err := s.fileWrite.WriteString(fmt.Sprintf("%d::%s::%s\n", id, value, userID))
	return err
}

func getLastIDOrDefault(file *os.File) (int64, error) {
	line, err := getLastLineOfTheFile(file)
	if err != nil {
		return 0, err
	}

	if len(line) == 0 {
		return 0, nil
	}

	id, _, _, ok := extractValuesFromTheLine(line)
	if ok {
		return id, err
	}
	return 0, fmt.Errorf("error while parsing last line of the fileWrite")

}
