package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

var (
	Extension  string
	ThumbWidth string
	MimeType   string
)

func init() {
	Extension = os.Getenv("EXTENSION")
	ThumbWidth = os.Getenv("THUMB_WIDTH")
	MimeType = os.Getenv("MIME_TYPE")
}

// slice string to bytes
func slice(s string) (b []byte) {
	pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pString := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pBytes.Data = pString.Data
	pBytes.Len = pString.Len
	pBytes.Cap = pString.Len
	return
}

// md5Str returns a string's md5
func md5Str(str string) string {
	h := md5.New()
	h.Write(slice(str))
	return hex.EncodeToString(h.Sum(nil))
}

// download s3 object and save to local file
func download(ctx context.Context, s3Client *s3.Client, bucket, key, filename string) error {
	object, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucket, key, err)
		return err
	}
	defer object.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", filename, err)
		return err
	}
	defer file.Close()

	body, err := io.ReadAll(object.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v:%v. Here's why: %v\n", bucket, key, err)
	}
	_, err = file.Write(body)
	if err != nil {
		log.Printf("Couldn't write file to %v. Here's why: %v\n", filename, err)
	}
	return err
}

// upload local file to s3 bucket
func upload(ctx context.Context, s3Client *s3.Client, bucket, key, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Couldn't open file %v to upload. Here's why: %v\n", filename, err)
		return err
	}
	defer file.Close()
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(MimeType),
	})
	if err != nil {
		log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n", file, bucket, key, err)
	}
	return err
}

func assembleResultKey(key string) string {
	if key == "" {
		return ""
	}
	copiedKey := key
	if strings.HasPrefix(key, "/") {
		copiedKey = strings.TrimLeft(key, "/")
	}
	re := regexp.MustCompile("\\.[^.]+$")
	// u/path/to/media.mp4 -> u/path/to/media.${Extension}
	// u/path/to/media -> u/path/to/media
	imageObjectKey := re.ReplaceAllString(copiedKey, fmt.Sprintf(".%s", Extension))
	objectSubPath := imageObjectKey[strings.Index(imageObjectKey, "/")+1:]
	// u/path/to/media.${Extension} -> thumb/path/to/media.${Extension}
	// u/path/to/media -> thumb/path/to/media
	return fmt.Sprintf("thumb/%s", objectSubPath)
}

func analyseObjectWithFFProbe(ctx context.Context, inputFile string) (*ffprobe.ProbeData, error) {
	ffprobe.SetFFProbeBinPath("/opt/bin/ffprobe")
	data, err := ffprobe.ProbeURL(ctx, inputFile)
	if err != nil {
		return nil, err
	}
	if data == nil || data.Format == nil || len(data.Streams) == 0 {
		return nil, errors.New(fmt.Sprintf("missing meta with file %s", inputFile))
	}
	return data, nil
}

func handler(ctx context.Context, s3Event events.S3Event) (string, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "cannot load default sdk config", err
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key
		md5 := md5Str(key)
		keyExt := path.Ext(key)
		resultKey := assembleResultKey(key)
		workDir := os.TempDir()
		inputFile := path.Join(workDir, fmt.Sprintf("%s%s", md5, keyExt))
		outputFile := path.Join(workDir, fmt.Sprintf("cover-%s.%s", md5, Extension))

		log.Printf("record: %#v", record)
		log.Printf("bucket: %s | key: %s | md5: %s | resultKey: %s | input: %s | output: %s\n", bucket, key, md5, resultKey, inputFile, outputFile)

		// download object
		if err0 := download(ctx, s3Client, bucket, key, inputFile); err0 != nil {
			return fmt.Sprintf("error download object %v:%v", bucket, key), err0
		}

		// checking media info
		probeData, err0 := analyseObjectWithFFProbe(ctx, inputFile)
		if err0 != nil {
			return fmt.Sprintf("error get object info by ffprobe with file %s", inputFile), err0
		}
		if probeData == nil || len(probeData.Streams) == 0 {
			return fmt.Sprintf("missing media probe info with file %s, returns", inputFile), nil
		}
		hitVideo := false
		for _, stream := range probeData.Streams {
			if stream.CodecType != string(ffprobe.StreamVideo) {
				continue
			}
			hitVideo = true
			if stream.Width+stream.Height <= 0 {
				return fmt.Sprintf("file %s is not an acceptable media file, returns", inputFile), nil
			}
		}
		if !hitVideo {
			return fmt.Sprintf("no video/image content in file %s, returns", inputFile), nil
		}

		// generate cover
		cmd := exec.Command("/opt/bin/ffmpeg", "-loglevel", "error", "-y", "-i", inputFile, "-frames:v", "1", "-vf", fmt.Sprintf("scale=%s:%s/a", ThumbWidth, ThumbWidth), outputFile)
		if err1 := cmd.Run(); err1 != nil {
			log.Printf("Couldn't get cover from file %v. Here's why: %v\n", inputFile, err1)
			return "", err1
		}
		log.Printf("cmd: %s\n", cmd.String())

		// upload cover
		if err2 := upload(ctx, s3Client, bucket, resultKey, outputFile); err2 != nil {
			return fmt.Sprintf("error upload file %v", outputFile), err2
		}
	}
	return "ok", nil
}

func main() {
	lambda.Start(handler)
}
