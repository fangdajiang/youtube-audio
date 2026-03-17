package handler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/wader/goutubedl"
	"google.golang.org/api/option"
	youtubeapi "google.golang.org/api/youtube/v3"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"youtube-audio/pkg/reporter"
	"youtube-audio/pkg/util"
	"youtube-audio/pkg/util/env"
	"youtube-audio/pkg/util/log"
	io2 "youtube-audio/pkg/util/myio"
	"youtube-audio/pkg/util/resource"
)

const (
	QualityScaleFrom0To9 string = "6"
)

type YouTubeCredentials struct {
	Key string
}

// AudioFormatInfo 保存音频格式筛选所需的信息
type AudioFormatInfo struct {
	FormatID       string
	ABR            float64 // 平均音频比特率 (KBit/s)
	Filesize       float64 // 文件大小（字节）
	FilesizeApprox float64 // 估计文件大小（字节）
	ACodec         string
}

type PlaylistVideoMetaData struct {
	Artist   string
	Album    string
	VideoId  string
	RawUrl   string
	Position int64
}
type PlaylistMetaData struct {
	PlaylistId                 string
	PlaylistVideoMetaDataArray []*PlaylistVideoMetaData
}

func (s PlaylistMetaData) Len() int {
	return len(s.PlaylistVideoMetaDataArray)
}
func (s PlaylistMetaData) Less(i, j int) bool {
	return s.PlaylistVideoMetaDataArray[i].Position > s.PlaylistVideoMetaDataArray[j].Position
}
func (s PlaylistMetaData) Swap(i, j int) {
	s.PlaylistVideoMetaDataArray[i], s.PlaylistVideoMetaDataArray[j] = s.PlaylistVideoMetaDataArray[j], s.PlaylistVideoMetaDataArray[i]
}

func FlushFetchHistory(deliveries []Delivery) {
	var fetchHistory = resource.FetchHistory{Playlists: GenerateFetchHistory(deliveries)}
	log.Debugf("fetchHistory: %v", fetchHistory)

	resource.MarshalPlaylistJson(fetchHistory)
}

func ProcessOneVideo(delivery *Delivery) {
	if !delivery.Done {
		audioFile, err := fetchAudio(delivery)
		if err != nil {
			log.Warnf("Failed to download audio url %s from YouTube, error: %v", delivery.Parcel.Url, err)
			SendWarningMessage(util.FailedToFetchAudioWarningTemplate, err.Error())
			return
		}
		if v, template := IsAudioValid(audioFile); v == false {
			log.Warnf("Downloaded file from YouTube %s is NOT valid: %s", delivery.Parcel.Url, template)
			SendWarningMessage(template, audioFile.FilePath)
			return
		} else {
			delivery.Parcel = audioFile
			err = SendAudio(delivery)

			if err != nil {
				log.Warnf("Failed to send file %s to telegram channel, error: %v", audioFile.FilePath, err)
				audioFile.Caption = audioFile.Caption + fmt.Sprintf("%s", err)
				SendWarningMessage(util.FailedToSendAudioWarningTemplate, audioFile.Caption)
			} else {
				reporter.BriefSummary.SuccessfulFetch++
			}
		}
		if audioFile.FilePath != "" {
			log.Debugf("ready to clean up, audioFile: %v", audioFile)
			io2.Cleanup(audioFile.FilePath)
		}
	} else {
		log.Infof("the state of this delivery %v is DONE, no more process", delivery)
	}
	fmt.Println()
}

func fetchAudio(delivery *Delivery) (Parcel, error) {
	// download a video
	return DownloadYouTubeAudioToPath(delivery)
}

func GetYouTubeService() (*youtubeapi.Service, error) {
	youTubeCredentials, err := GenerateYouTubeCredentials()
	if err != nil {
		log.Errorf("generate youtube credentials error:%v", err)
		return nil, err
	}

	ctx := context.Background()
	svc, err := youtubeapi.NewService(ctx, option.WithScopes(youtubeapi.YoutubeReadonlyScope), option.WithAPIKey(youTubeCredentials.Key))
	if err != nil {
		log.Errorf("new service error:%v", err)
		return nil, err
	}

	return svc, nil
}

func GetPlaylistMetaDataBy(playlistId string) PlaylistMetaData {
	var playlistMetaData PlaylistMetaData
	svc, err := GetYouTubeService()
	if err != nil {
		log.Errorf("get youtube service error:%v", err)
		return playlistMetaData
	}

	playlistResponse := playlistItemsList(svc, util.YouTubePart, playlistId)

	playlistMetaData.PlaylistId = playlistId
	for _, playlistItem := range playlistResponse.Items {
		publishedAt := playlistItem.Snippet.PublishedAt
		title := playlistItem.Snippet.Title
		localPublishedAt := util.GetLocalDateTime(publishedAt)
		channelTitle := playlistItem.Snippet.ChannelTitle
		channelId := playlistItem.Snippet.ChannelId

		videoId := playlistItem.Snippet.ResourceId.VideoId
		position := playlistItem.Snippet.Position
		log.Debugf("%s(%s) from %s(%s) on position %v artist(%s/%s) was published at %s",
			title, videoId, channelTitle, channelId, position,
			util.GetYouTubePlaylistArtist(playlistId),
			util.GetYouTubePlaylistAlbum(playlistId),
			localPublishedAt)

		videoMetaData := PlaylistVideoMetaData{
			util.GetYouTubePlaylistArtist(playlistId),
			util.GetYouTubePlaylistAlbum(playlistId),
			videoId,
			util.MakeYouTubeRawUrl(videoId),
			position}
		playlistMetaData.PlaylistVideoMetaDataArray = append(playlistMetaData.PlaylistVideoMetaDataArray, &videoMetaData)
	}

	return playlistMetaData
}

func playlistItemsList(service *youtubeapi.Service, part []string, playlistId string) *youtubeapi.PlaylistItemListResponse {
	call := service.PlaylistItems.List(part)
	call = call.PlaylistId(playlistId)
	call = call.MaxResults(util.GetYouTubePlaylistMaxResultsCount(playlistId))
	response, err := call.Do()
	if err != nil {
		log.Errorf("get playlist items error:%v, playlistId:%s", err, playlistId)
	}
	return response
}

// retrieveAudioFormatIDs 从 goutubedl 返回的格式列表中筛选纯音频格式，
// 按文件大小升序排列，返回格式 ID 列表（最小尺寸优先）
func retrieveAudioFormatIDs(formats []goutubedl.Format) []string {
	var audioFormats []AudioFormatInfo
	for _, f := range formats {
		log.Debugf("FormatID:%s, Ext:%s, ACodec:%s, VCodec:%s, ABR:%.0f, FPS:%.0f, Filesize:%.0f, FilesizeApprox:%.0f",
			f.FormatID, f.Ext, f.ACodec, f.VCodec, f.ABR, f.FPS, f.Filesize, f.FilesizeApprox)
		// 筛选纯音频格式：无视频编解码器，且有音频编解码器
		if (f.VCodec == "none" || f.VCodec == "") && f.ACodec != "none" && f.ACodec != "" {
			filesize := f.Filesize
			if filesize == 0 {
				filesize = f.FilesizeApprox
			}
			audioFormats = append(audioFormats, AudioFormatInfo{
				FormatID:       f.FormatID,
				ABR:            f.ABR,
				Filesize:       f.Filesize,
				FilesizeApprox: f.FilesizeApprox,
				ACodec:         f.ACodec,
			})
		}
	}
	if len(audioFormats) == 0 {
		log.Warnf("no audio-only formats found among %d total formats", len(formats))
		return nil
	}

	// 按文件大小升序排列（最小文件优先）
	sort.Slice(audioFormats, func(i, j int) bool {
		sizeI := audioFormats[i].Filesize
		if sizeI == 0 {
			sizeI = audioFormats[i].FilesizeApprox
		}
		if sizeI == 0 {
			sizeI = audioFormats[i].ABR // 无文件大小时用比特率近似
		}
		sizeJ := audioFormats[j].Filesize
		if sizeJ == 0 {
			sizeJ = audioFormats[j].FilesizeApprox
		}
		if sizeJ == 0 {
			sizeJ = audioFormats[j].ABR
		}
		return sizeI < sizeJ
	})

	// 提取格式 ID 列表
	var formatIDs []string
	for _, af := range audioFormats {
		log.Debugf("audio format: ID=%s, ACodec=%s, ABR=%.0f, Filesize=%.0f, FilesizeApprox=%.0f",
			af.FormatID, af.ACodec, af.ABR, af.Filesize, af.FilesizeApprox)
		formatIDs = append(formatIDs, af.FormatID)
	}

	return formatIDs
}

func DownloadYouTubeAudioToPath(delivery *Delivery) (Parcel, error) {
	var parcel Parcel
	log.Debugf("Ready to download media %s(playlistId: %s) at %s", delivery.Parcel.Url, delivery.PlaylistId, time.Now().Format(util.DateTimeFormat))
	result, err := goutubedl.New(context.Background(), delivery.Parcel.Url, goutubedl.Options{DownloadThumbnail: true})
	if err != nil {
		log.Errorf("goutubedl error:%s", err)
		return parcel, fmt.Errorf("goutubedl new error: %v, url: %s", err, delivery.Parcel.Url)
	}

	fileExtension := getFileExtension(result.Info.ACodec)
	validMediaFileName := util.FilenamifyMediaTitle(result.Info.Title + fileExtension)
	parcelFilePath := fmt.Sprintf("%s%s", util.GetYouTubeFetchBase().DownloadedFilesPath, validMediaFileName)
	thumbnailBytes, thumbErr := normalizeThumbnail(result.Info.ThumbnailBytes)
	if thumbErr != nil {
		log.Warnf("normalize thumbnail error: %v", thumbErr)
		thumbnailBytes = nil
	}
	log.Debugf("ext: %s, parcelFilePath: %s, thumbnailBytes: %v, result.Info.ACodec: %s",
		fileExtension, parcelFilePath, len(thumbnailBytes), result.Info.ACodec)
	parcel = GenerateParcel(
		parcelFilePath,
		result.Info.Title,
		util.GetYouTubePlaylistArtist(delivery.PlaylistId),
		util.GetYouTubePlaylistAlbum(delivery.PlaylistId),
		delivery.Parcel.Url,
		result.Info.Duration,
		thumbnailBytes,
		result.Info.FilesizeApprox)
	log.Debugf("ext: %s, title: %s, artist: %s, album: %s, url: %s, duration: %v, thumbnailBytes: %v, filesizeApprox: %v",
		fileExtension, parcel.Caption, parcel.Artist, parcel.Album, parcel.Url,
		parcel.Duration, len(parcel.ThumbnailBytes), parcel.FilesizeApprox)

	log.Debugf("ready to CREATE media file %s at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	parcelFile, err := os.Create(parcel.FilePath)
	log.Debugf("media file %s CREATED at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	if err != nil {
		log.Errorf("creating file error: %v", err)
		return parcel, fmt.Errorf("creating file error: %v", err)
	}

	// 从 goutubedl 获取的格式列表中筛选音频格式
	formatIDs := retrieveAudioFormatIDs(result.Info.Formats)
	log.Infof("found %d audio format(s) for url: %s", len(formatIDs), delivery.Parcel.Url)

	var downloadedResult *goutubedl.DownloadResult
	// 按文件大小从小到大尝试下载
	for _, formatID := range formatIDs {
		downloadedResult, err = result.Download(context.Background(), formatID)
		if err == nil {
			log.Debugf("successfully selected formatID=%s at %s", formatID, time.Now().Format(util.DateTimeFormat))
			break
		}
		log.Errorf("download error with formatID %s: %s", formatID, err)
	}

	// Fallback：如果所有格式都失败，让 yt-dlp 选择最小体积的音频
	if err != nil {
		log.Warnf("all format IDs failed, falling back to yt-dlp worstaudio mode for url: %s", delivery.Parcel.Url)
		downloadedResult, err = result.Download(context.Background(), "worstaudio")
	}

	if err != nil {
		log.Errorf("final download error:%s", err)
		return parcel, fmt.Errorf("goutubedl download error: %v, url: %s", err, delivery.Parcel.Url)
	}
	defer func(downloadedResult *goutubedl.DownloadResult) {
		_ = downloadedResult.Close()
	}(downloadedResult)
	log.Debugf("downloading media %s at %s", result.Info.Title, time.Now().Format(util.DateTimeFormat))

	log.Debugf("ready to COPY media file %s at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	written, err := io.Copy(parcelFile, downloadedResult)
	log.Infof("media file %s DOWNLOADED & COPIED at %s", parcel.FilePath, time.Now().Format(util.DateTimeFormat))
	if err != nil {
		_ = parcelFile.Close()
		_ = os.Remove(parcel.FilePath)
		return parcel, fmt.Errorf("copy error: %s, parcel: %v, written: %v", err, parcel, written)
	}
	if err := parcelFile.Close(); err != nil {
		log.Warnf("closing downloaded file error: %v", err)
	}

	log.Printf("Title: %s, Artist: %s, Album: %s, Url: %s", parcel.Caption, parcel.Artist, parcel.Album, parcel.Url)
	parcel, err = convertToMp3AndFillMetadata(parcel)
	if err != nil {
		_ = os.Remove(parcel.FilePath)
		return Parcel{}, fmt.Errorf("convert to mp3 error: %w", err)
	}

	return parcel, nil
}
func convertToMp3AndFillMetadata(parcel Parcel) (Parcel, error) {
	// 生成新的文件路径，使用.mp3作为扩展名
	originalFilePath := parcel.FilePath
	newFilePath := strings.TrimSuffix(parcel.FilePath, filepath.Ext(parcel.FilePath)) + ".mp3"
	var coverTempPath string
	if len(parcel.ThumbnailBytes) > 0 {
		tempFile, err := os.CreateTemp("", "ya-cover-*.png")
		if err != nil {
			log.Warnf("create temp cover file error: %v", err)
		} else {
			writeOK := true
			if _, err := tempFile.Write(parcel.ThumbnailBytes); err != nil {
				log.Warnf("write temp cover file error: %v", err)
				writeOK = false
			}
			if err := tempFile.Close(); err != nil {
				log.Warnf("close temp cover file error: %v", err)
				writeOK = false
			}
			if writeOK {
				coverTempPath = tempFile.Name()
				defer func(name string) {
					_ = os.Remove(name)
				}(coverTempPath)
			} else {
				_ = os.Remove(tempFile.Name())
			}
		}
	}

	// 构建ffmpeg命令
	args := []string{"-y", "-i", parcel.FilePath}
	baseMetadata := []string{
		"-metadata", "artist=" + parcel.Artist,
		"-metadata", "title=" + util.FilenamifyMediaTitle(parcel.Caption),
		"-metadata", "album=" + parcel.Album,
		"-id3v2_version", "3",
	}
	if coverTempPath != "" {
		args = append(args,
			"-i", coverTempPath,
			"-map", "0:a",
			"-map", "1:v",
			"-codec:a", "libmp3lame",
			"-qscale:a", QualityScaleFrom0To9,
			"-codec:v", "png",
			"-disposition:v:0", "attached_pic",
			"-metadata:s:v", "title=Album cover",
			"-metadata:s:v", "comment=Cover (front)",
		)
	} else {
		args = append(args,
			"-codec:a", "libmp3lame",
			"-qscale:a", QualityScaleFrom0To9,
		)
	}
	args = append(args, baseMetadata...)
	args = append(args, newFilePath)

	cmd := exec.Command("ffmpeg", args...)

	fullCommand := strings.Join(cmd.Args, " ")
	log.Printf("Executing command: %s", fullCommand)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("ffmpeg error: %s, stderr: %s", err, stderr.String())
		_ = os.Remove(newFilePath)
		return parcel, fmt.Errorf("ffmpeg error: %s, command: %s", err, fullCommand)
	}

	ffprobeCmd := exec.Command("ffprobe", "-show_format", newFilePath)
	output, err := ffprobeCmd.CombinedOutput()
	if err != nil {
		log.Errorf("ffprobe error: %s", err)
		_ = os.Remove(newFilePath)
		return parcel, fmt.Errorf("ffprobe error: %s", err)
	}
	log.Printf("ffprobe output: %s", string(output))

	parcel.FilePath = newFilePath
	if err := os.Remove(originalFilePath); err != nil && !os.IsNotExist(err) {
		log.Warnf("remove original media error: %v", err)
	}
	fileInfo, err := os.Stat(newFilePath)
	if err != nil {
		log.Warnf("stat converted file error: %v", err)
	} else {
		parcel.FilesizeApprox = float64(fileInfo.Size())
	}
	log.Printf("ffmpeg command executed successfully, new file: %s", parcel.FilePath)

	return parcel, nil
}
func getFileExtension(mimeType string) string {
	// 简单的MimeType到文件扩展名的映射
	switch {
	case strings.Contains(mimeType, "audio/mp4"):
		return ".m4a"
	case strings.Contains(mimeType, "audio/webm"):
		return ".webm"
	case strings.Contains(mimeType, "audio/ogg"):
		return ".ogg"
	default:
		return ".mp4" // 默认情况下使用.mp4，适用于不确定的情况
	}
}

func normalizeThumbnail(thumbnail []byte) ([]byte, error) {
	if len(thumbnail) == 0 {
		return nil, nil
	}

	contentType := http.DetectContentType(thumbnail)
	if contentType == "image/jpeg" || contentType == "image/png" {
		return thumbnail, nil
	}

	srcFile, err := os.CreateTemp("", "ya-thumb-src-*")
	if err != nil {
		return nil, fmt.Errorf("create temp thumbnail source error: %w", err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(srcFile.Name())

	if _, err = srcFile.Write(thumbnail); err != nil {
		_ = srcFile.Close()
		return nil, fmt.Errorf("write temp thumbnail source error: %w", err)
	}
	if err = srcFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp thumbnail source error: %w", err)
	}

	dstFile, err := os.CreateTemp("", "ya-thumb-out-*.png")
	if err != nil {
		return nil, fmt.Errorf("create temp thumbnail output error: %w", err)
	}
	dstPath := dstFile.Name()
	_ = dstFile.Close()
	defer func(name string) {
		_ = os.Remove(name)
	}(dstPath)

	cmd := exec.Command("ffmpeg", "-y", "-i", srcFile.Name(), dstPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg thumbnail conversion error: %w, stderr: %s", err, stderr.String())
	}

	converted, err := os.ReadFile(dstPath)
	if err != nil {
		return nil, fmt.Errorf("read converted thumbnail error: %w", err)
	}
	if len(converted) == 0 {
		return nil, fmt.Errorf("converted thumbnail is empty")
	}

	return converted, nil
}

func GenerateFetchHistory(deliveries []Delivery) []resource.HistoryProps {
	channelChatId, _ := env.GetEnvVariable(util.EnvChatIdName)
	subscriberId, _ := strconv.ParseInt(channelChatId, 10, 64)

	var playlistMap map[string][]resource.SubscriberItems
	playlistMap = make(map[string][]resource.SubscriberItems)
	for _, delivery := range deliveries {
		log.Debugf("delivery to be generated: %v", delivery)
		subscribers, ok := playlistMap[delivery.PlaylistId]
		if ok {
			log.Infof("playlist id %s FOUND from playlistMap, length: %v", delivery.PlaylistId, len(playlistMap))
			var newSubscribers []resource.SubscriberItems
			if delivery.Done {
				for _, sub := range subscribers {
					if sub.Id == subscriberId {
						sub.LastFetch = resource.FetchItems{Datetime: sub.LastFetch.Datetime, Timestamp: sub.LastFetch.Timestamp, Urls: append(sub.LastFetch.Urls, delivery.Parcel.Url)}
						newSubscribers = append(newSubscribers, sub)
					}
				}
			} else {
				log.Debugf("delivery.done FALSE: %v", delivery)
				now := time.Now()
				for _, sub := range subscribers {
					log.Debugf("sub: %v", sub)
					nextFetchTimestamp := delivery.Timestamp
					nextFetchDatetime := delivery.Datetime
					if nextFetchTimestamp == 0 {
						nextFetchTimestamp = now.Unix()
						nextFetchDatetime = now.Format(util.DateTimeFormat)
					}
					if sub.Id == subscriberId {
						sub.NextFetch = resource.FetchItems{Datetime: nextFetchDatetime, Timestamp: nextFetchTimestamp, Urls: append(sub.NextFetch.Urls, delivery.Parcel.Url)}
						newSubscribers = append(newSubscribers, sub)
					}
				}
			}
			log.Debugf("newSubscribers: %v", newSubscribers)
			playlistMap[delivery.PlaylistId] = newSubscribers
		} else {
			log.Infof("playlist id %s NOT FOUND from playlistMap: %v", delivery.PlaylistId, playlistMap)
			var urls []string
			urls = append(urls, delivery.Parcel.Url)
			var lastFetch, nextFetch resource.FetchItems
			thisFetch := resource.FetchItems{Datetime: delivery.Datetime, Timestamp: delivery.Timestamp, Urls: urls}
			if delivery.Done {
				lastFetch = thisFetch
				log.Debugf("lastFetch: %v", lastFetch)
			} else {
				nextFetch = thisFetch
				log.Debugf("nextFetch: %v", nextFetch)
			}
			subscriberItem := resource.SubscriberItems{Id: subscriberId, LastFetch: lastFetch, NextFetch: nextFetch}
			subscriberItems := []resource.SubscriberItems{subscriberItem}
			playlistMap[delivery.PlaylistId] = subscriberItems
		}
	}
	log.Debugf("playlistMap: %v", playlistMap)
	var historyPropsArray []resource.HistoryProps
	for playlistId := range playlistMap {
		historyProps := resource.HistoryProps{Id: playlistId, Subscribers: playlistMap[playlistId]}
		historyPropsArray = append(historyPropsArray, historyProps)
	}
	return historyPropsArray
}

func MergeHistoryFetchesInto(newDeliveries []Delivery) []Delivery {
	historyProps := resource.MediaHistory
	log.Debugf("newDeliveries count: %v, historyProps count: %v", len(newDeliveries), len(historyProps))
	var mergedDeliveries []Delivery
	for _, newDelivery := range newDeliveries {
		log.Debugf("newDelivery: %v", newDelivery)
		isNewPlayListId := true
		for _, historyProp := range historyProps {
			if newDelivery.PlaylistId == historyProp.Id {
				isNewPlayListId = false
				for _, sub := range historyProp.Subscribers {
					AppendDeliveries(&mergedDeliveries, sub.LastFetch, historyProp.Id, true)
					nextFetchUrls := sub.NextFetch.Urls
					if len(nextFetchUrls) > 0 {
						if util.StringSliceContains(nextFetchUrls, newDelivery.Parcel.Url) {
							log.Infof("newDelivery url %s was FOUND from history NEXT fetch urls: %v", newDelivery.Parcel.Url, nextFetchUrls)
							AppendDeliveries(&mergedDeliveries, sub.NextFetch, historyProp.Id, false)
						} else {
							log.Infof("newDelivery url %s NOT FOUND from history NEXT fetch urls: %v, add it", newDelivery.Parcel.Url, nextFetchUrls)
							mergedDeliveries = append(mergedDeliveries, newDelivery)
						}
					} else {
						log.Infof("next fetch urls EMPTY, subscribers id: %v, playlist id: %v, add it", sub.Id, historyProp.Id)
						mergedDeliveries = append(mergedDeliveries, newDelivery)
					}
				}
				break
			}
		}
		if isNewPlayListId {
			now := time.Now()
			newDelivery.Timestamp = now.Unix()
			newDelivery.Datetime = now.Format(util.DateTimeFormat)
			mergedDeliveries = append(mergedDeliveries, newDelivery)
		}
	}
	mergedDeliveriesWithoutDuplicated := RemoveDuplicatedUrlsByLoop(mergedDeliveries)
	log.Infof("merged deliveries count: %v which removed duplicated items", len(mergedDeliveriesWithoutDuplicated))
	return mergedDeliveriesWithoutDuplicated
}

func GetYouTubeVideosFromPlaylistId(playlistId string) []PlaylistMetaData {
	return []PlaylistMetaData{GetPlaylistMetaDataBy(playlistId)}
}

func GetYouTubeVideosFromPlaylists() []PlaylistMetaData {
	var playlistMetaDataArray []PlaylistMetaData
	for _, param := range util.GetYouTubeFetchBase().Params {
		playlistMetaData := GetPlaylistMetaDataBy(param.Id)
		if param.SortByPosition {
			log.Debugf("SORT the playlist:%s", param.Id)
			sort.Sort(playlistMetaData)
		}
		playlistMetaDataArray = append(playlistMetaDataArray, playlistMetaData)
	}
	return playlistMetaDataArray
}

func AssembleDeliveryFromSingleUrl(url string) Delivery {
	parcel := Parcel{Url: url}
	return Delivery{Parcel: parcel}
}

func AssembleDeliveriesFromPlaylists(playlistMetaDataArray []PlaylistMetaData) []Delivery {
	var deliveries []Delivery
	for _, playlistMetaData := range playlistMetaDataArray {
		delivery := Delivery{}
		delivery.PlaylistId = playlistMetaData.PlaylistId
		for _, playlistVideoMetaData := range playlistMetaData.PlaylistVideoMetaDataArray {
			delivery.Parcel = GenerateParcel("", "", playlistVideoMetaData.Artist,
				playlistVideoMetaData.Album, playlistVideoMetaData.RawUrl, 0.0, nil, 0.0)
			deliveries = append(deliveries, delivery)
		}
	}
	log.Infof("total incoming playlists: %v, total incoming deliveries: %v", len(playlistMetaDataArray), len(deliveries))
	return deliveries
}

func GenerateYouTubeCredentials() (YouTubeCredentials, error) {
	var err error
	var youTubeCredentials YouTubeCredentials

	youtubeKey, err := env.GetEnvVariable(util.EnvYouTubeKeyName)
	if err != nil {
		log.Errorf("%s", err)
		return youTubeCredentials, fmt.Errorf("reading env %s vars error", util.EnvYouTubeKeyName)
	}
	youTubeCredentials.Key = youtubeKey

	return youTubeCredentials, nil
}
