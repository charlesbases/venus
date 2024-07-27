#!/usr/bin/env bash

# 视频地址
# $1="https://surrit.com/2ce795b0-70da-49e0-aa84-3976fc3d7423/1080p/video.m3u8"
remote=$1


# no need to modify
root="missav.com"
output="$root/$(date +'%Y%m%d%H%M%S').mp4"

header_user_agent="User-Agent:Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36"


# ready
if [ ! -d "$root" ]; then
  mkdir $root
fi


# page
page=$(curl -s -H "$header_user_agent" $remote | grep -- ".jpeg" | tail -n 1 | sed 's/video\([0-9]*\).jpeg/\1/')

if [ -z $page ]; then
  echo -e "\033[31minvalid address\033[0m"
  exit
fi

# url
url="${remote%/*}"


# download
seq 0 $page | while read n; do
  echo -e " $output [$n/$page]"
  echo -e '\033[1A\033[1K\c'

  curl -s -m 3600 -H "$header_user_agent" "$url/video$n.jpeg" >> $output
done

echo
