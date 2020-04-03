# otter_hls_server
### 简介：

使用Go实现的HLS协议的流媒体服务器，通过解析ts媒体文件建立索引，在请求时从源媒体文件实时生成ts流。

开发环境：golang 1.13.4 



### 配置文件：

```yaml
server:
  port: 4000
path:
  index_file_folder: /index/
  media_file_folders:
    - local_path: /var/media1/
      group_name: mediaPath1
    - local_path: /var/media2/
      group_name: mediaPath2
m3u8:
  target_duration: 10
log:
  syslog:
    filename: /var/log/otter_hls_server/system
    pattern: -2006-01-02
    level: debug
```

| 配置项                                | 描述                                       |
| ------------------------------------- | ------------------------------------------ |
| server.port                           | 服务监听端口                               |
| path.index_file_folder                | 索引文件保存根目录                         |
| path.media_file_folders               | 媒体文件目录列表，支持多目录配置           |
| path.media_file_folders[i].local_path | 媒体文件目录本地路径                       |
| path.media_file_folders[i].group_name | 媒体文件目录分组名（在请求m3u8路径中使用） |
| m3u8.targe_duration                   | m3u8 最大分片时长（单位秒）                |
| log.syslog.filename                   | 日志路径                                   |
| log.syslog.pattern                    | 日期分割表达式                             |
| log.syslog.level                      | 日志级别：debug、info、warn、error         |

### 路由：

#### /

欢迎页



#### /hls/{group_name}/xxx.m3u8

获取一级m3u8文件索引

例如:

​	媒体文件本地路径：/var/media2/demo/1.ts

​	配置文件:

```yaml
...
path:
  index_file_folder: /index/
  media_file_folders:
    - local_path: /var/media1/
      group_name: mediaPath1
    - local_path: /var/media2/
      group_name: mediaPath2
... 
```

则请求url为:

http://host:port/hls/mediaPath2/demo/1.m3u8

返回：

```m3u8
#EXTM3U
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=1164839
http://host:port/hls_sub/mediaPath2/demo/1.m3u8
```



#### /hls_sub/{group_name}/xxx.m3u8

获取二级m3u8文件索引

例如:

​	媒体文件本地路径：/var/media2/demo/1.ts

​	配置文件:

```yaml
...
path:
  index_file_folder: /index/
  media_file_folders:
    - local_path: /var/media1/
      group_name: mediaPath1
    - local_path: /var/media2/
      group_name: mediaPath2
... 
```

则请求url为:

http://host:port/hls_sub/mediaPath2/demo/1.m3u8

返回：

```
#EXTM3U
#EXT-X-VERSION:4 
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:VOD
#EXTINF:10.00
http://host:port/video/mediaPath2/demo/1_0.ts
#EXTINF:10.00
http://host:port/video/mediaPath2/demo/1_1.ts
```



#### /video/{group_name}/xxx_0.ts

获取媒体分片数据

例如:

​	媒体文件本地路径：/var/media2/demo/1.ts

​	配置文件:

```yaml
...
path:
  index_file_folder: /index/
  media_file_folders:
    - local_path: /var/media1/
      group_name: mediaPath1
    - local_path: /var/media2/
      group_name: mediaPath2
... 
```

则媒体第一个分片的请求url为:

http://host:port/video/mediaPath2/demo/1_0.ts

返回：

视频媒体文件 1_0.ts



#### /createIndex/{group_name}/xxx.ts

主动创建媒体文件索引

例如:

​	媒体文件本地路径：/var/media2/demo/1.ts

​	配置文件:

```yaml
...
path:
  index_file_folder: /index/
  media_file_folders:
    - local_path: /var/media1/
      group_name: mediaPath1
    - local_path: /var/media2/
      group_name: mediaPath2
... 
```

则请求URL为：

http://host:port/createIndex/mediaPath2/demo/1.ts

成功返回：

```json
{"code":"1","msg":""}
```

失败返回

```json
{"code":"-1","msg":"errMsg"}
```

