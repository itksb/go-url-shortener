package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/itksb/go-url-shortener/internal/config"
	"github.com/itksb/go-url-shortener/internal/handler"
	"github.com/itksb/go-url-shortener/internal/shortener"
	"github.com/itksb/go-url-shortener/internal/user"
	"github.com/itksb/go-url-shortener/pkg/logger"
	"github.com/itksb/go-url-shortener/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/url"
	"strings"
	"time"
)

var _ go_url_shortener.ShortenerServer = server{}

type server struct {
	// нужно встраивать тип go_url_shortener.UnimplementedShortenerServer
	// для совместимости с будущими версиями
	go_url_shortener.UnimplementedShortenerServer
	dbping       handler.IPingableDB
	logger       logger.Interface
	urlshortener *shortener.Service
	cfg          config.Config
}

func NewGRPCServer(
	dbping handler.IPingableDB,
	logger logger.Interface,
	urlshortener *shortener.Service,
	cfg config.Config,
) go_url_shortener.ShortenerServer {
	return &server{
		dbping:       dbping,
		logger:       logger,
		urlshortener: urlshortener,
		cfg:          cfg,
	}
}

func (s server) Ping(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	empty := &emptypb.Empty{}

	if !s.dbping.Ping(ctx2) {
		return empty, status.Error(codes.Unavailable, "Ping error")
	}
	return empty, nil
}

func (s server) ShortUrl(
	ctx context.Context,
	request *go_url_shortener.ShortURLRequest,
) (*go_url_shortener.ShortURLResponse, error) {

	result := &go_url_shortener.ShortURLResponse{}

	if request.GetUrl() == "" {
		s.logger.Error("ApiShortenUrl. Bad request. URL is empty")
		return nil, status.Error(codes.Unknown, "bad input request: URL is empty")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "couldn`t read meatadata")
	}

	if len(md.Get(string(user.FieldID))) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get(string(user.FieldID))[0]
	if userID == "" {
		s.logger.Error("no user id found")
		return nil, status.Error(codes.Unknown, "no user id found")
	}

	sURLId, err := s.urlshortener.ShortenURL(ctx, request.Url, userID)
	if err != nil && !errors.Is(err, shortener.ErrDuplicate) {
		s.logger.Error("ApiShortenUrl. urlshortener.ShortenURL(...) call error", err.Error())
		return nil, status.Error(codes.Unknown, "shortener service error")
	}

	result.Result = handler.CreateShortenURL(sURLId, s.cfg.ShortBaseURL)

	return result, nil
}

// Get returns original URL by short URL ID
// urlshortener.GetURL(r.Context(), id)
func (s server) GetURL(
	ctx context.Context,
	request *go_url_shortener.GetURLRequest,
) (*go_url_shortener.GetURLResponse, error) {

	parsedURL, err := url.Parse(request.GetUrl())
	if err != nil {
		s.logger.Info("url ID not found", request.GetUrl())
		return nil, status.Error(codes.InvalidArgument, "url id not found")
	}

	_, id, ok := strings.Cut(parsedURL.Path, "/")
	if !ok {
		s.logger.Error("parse url id")
		return nil, status.Error(codes.InvalidArgument, "url id not found")
	}

	listItem, err := s.urlshortener.GetURL(ctx, id)
	if err != nil {
		s.logger.Info("Id not found", id)
		return nil, status.Error(codes.NotFound, "id not found")
	}

	if len(listItem.OriginalURL) == 0 {
		s.logger.Info("Id not found:", id)
		return nil, status.Error(codes.NotFound, "id not found")
	}

	if listItem.DeletedAt != nil && *listItem.DeletedAt != "" {
		s.logger.Info("url is already deleted for id:", id)
		return nil, status.Error(codes.NotFound, "id not found")
	}
	//
	result := go_url_shortener.GetURLResponse{
		Result: listItem.OriginalURL,
	}

	return &result, nil
}

func (s server) ListUrl(
	ctx context.Context,
	empty *empty.Empty,
) (*go_url_shortener.ListURLResponse, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "couldn`t read meatadata")
	}

	if len(md.Get(string(user.FieldID))) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get(string(user.FieldID))[0]
	if userID == "" {
		s.logger.Error("no user id found")
		return nil, status.Error(codes.InvalidArgument, "no user id found")
	}

	urlListItems, err := s.urlshortener.ListURLByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("error while searching user urls")
		return nil, status.Error(codes.Unavailable, "shortener service error")
	}

	// creating short urls is infrastructure layer responsibility, that`s why it is here
	for idx := range urlListItems {
		urlListItems[idx].ShortURL = handler.CreateShortenURL(
			fmt.Sprint(urlListItems[idx].ID),
			s.cfg.ShortBaseURL,
		)
	}

	result := &go_url_shortener.ListURLResponse{
		Urls: make(
			[]*go_url_shortener.ListURLResponse_URLListItem,
			0,
			len(urlListItems),
		),
	}

	for _, item := range urlListItems {
		result.Urls = append(result.Urls, &go_url_shortener.ListURLResponse_URLListItem{
			OriginalUrl: item.OriginalURL,
			ShortUrl:    item.ShortURL,
		})
	}

	if len(urlListItems) > 0 {
		return result, nil
	} else {
		return nil, status.Error(codes.NotFound, "no urls found")
	}

}

func (s server) ShortenBatch(
	ctx context.Context,
	request *go_url_shortener.ShortURLBatchRequest,
) (*go_url_shortener.ShortURLBatchResponse, error) {
	if len(request.Urls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no urls provided")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "couldn`t read meatadata")
	}

	if len(md.Get(string(user.FieldID))) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get(string(user.FieldID))[0]
	if userID == "" {
		s.logger.Error("no user id found")
		return nil, status.Error(codes.InvalidArgument, "no user id found")
	}

	response := go_url_shortener.ShortURLBatchResponse{
		Urls: make([]*go_url_shortener.ShortURLBatchResponse_URLItem, 0, len(request.Urls)),
	}
	conflict := false

	for _, shortenBatchItemRequest := range request.Urls {
		sURLId, err := s.urlshortener.ShortenURL(ctx, shortenBatchItemRequest.OriginalUrl, userID)
		if err != nil && !errors.Is(err, shortener.ErrDuplicate) {
			s.logger.Error("ApiShortenUrl. urlshortener.ShortenURL(...) call error", err.Error())
			return nil, status.Error(codes.Unavailable, "shortener service error")
		}
		if errors.Is(err, shortener.ErrDuplicate) {
			conflict = true
		}
		shortURL := handler.CreateShortenURL(sURLId, s.cfg.ShortBaseURL)
		responseItem := &go_url_shortener.ShortURLBatchResponse_URLItem{
			CorrelationId: shortenBatchItemRequest.CorrelationId,
			ShortUrl:      shortURL,
		}
		response.Urls = append(response.Urls, responseItem)
	}

	if conflict {
		return &response, status.Error(codes.AlreadyExists, "some urls already exists")
	} else {
		return &response, nil
	}
}

func (s server) DeleteBatch(
	ctx context.Context,
	request *go_url_shortener.DeleteURLBatchRequest,
) (*empty.Empty, error) {

	if len(request.Ids) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no ids provided")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "couldn`t read meatadata")
	}

	if len(md.Get(string(user.FieldID))) == 0 {
		return nil, status.Error(codes.Unknown, "wrong metadata")
	}

	userID := md.Get(string(user.FieldID))[0]
	if userID == "" {
		s.logger.Error("no user id found")
		return nil, status.Error(codes.InvalidArgument, "no user id found")
	}

	err := s.urlshortener.DeleteURLBatch(ctx, userID, request.Ids)
	if err != nil {
		s.logger.Error(fmt.Sprintf("error while DeleteURLBatch: %s", err.Error()))
		return nil, status.Error(codes.Unavailable, "shortener service error")
	}

	return &empty.Empty{}, nil
}

func (s server) InternalStats(ctx context.Context, empty *empty.Empty) (*go_url_shortener.InternalStatsResponse, error) {
	stats, err := s.urlshortener.GetStatistics(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("error while InternalStats: %s", err.Error()))
		return nil, status.Error(codes.Unavailable, "shortener service error")
	}

	return &go_url_shortener.InternalStatsResponse{
		Urls:  uint64(stats.URLs),
		Users: uint64(stats.Users),
	}, nil
}
