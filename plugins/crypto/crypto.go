package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/AyakuraYuki/s3-video-cover-builder/plugins/hack"
)

func Md5Str(str string) string {
	h := md5.New()
	h.Write(hack.Slice(str))
	return hex.EncodeToString(h.Sum(nil))
}
