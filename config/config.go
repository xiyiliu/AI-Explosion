package config

var Secret = "tiktok"

var OneDayOfHours = 60 * 60 * 24

// VideoCount 每次获取视频流的数量
const VideoCount = 5

// PlayUrlPrefix 存储的图片和视频的链接
const PlayUrlPrefix = "http://47.113.148.197/" //服务器地址
const CoverUrlPrefix = "http://47.113.148.197/images/"

// ConConfig ftp服务器地址
const ConConfig = "47.113.148.197:21"
const FtpUser = "ftpuser"
const FtpPsw = "123456"
const HeartbeatTime = 2 * 60
