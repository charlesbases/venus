#!/usr/bin/env bash

set -e

mkdir -p resp

# 用户首页的视频列表
curl https://www.xvideos.com/channels/chicken1806/videos/new/0 > resp/user.json

# 根据视频网页地址，获取 hls 地址
#curl https://www.xvideos.com/video75180633/_ > resp/video.html
