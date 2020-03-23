# otter_hls_server
HLS协议流媒体服务器

1、不真的拆解ts文件为多个分片ts

2、通过建立ts索引，在请求时从源媒体文件生成ts流

golang 1.13.4 