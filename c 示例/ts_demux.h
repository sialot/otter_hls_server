#pragma once
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "byte_list.h"
#include "hash_map.h"
#include "frame_data.h"

// ts 头
typedef struct TS_HEADER
{
	unsigned sync_byte                       : 8;  // 同步字节：固定为0x47;
	unsigned transport_error_indicator       : 1;  // 传输错误标志：‘1’表示在相关的传输包中至少有一个不可纠正的错误位。
	unsigned payload_unit_start_indicator    : 1;  // 负载起始标志：在前4个字节之后会有一个调整字节，其的数值为后面调整字段的长度length。
	unsigned transport_priority              : 1;  // 传输优先级标志
	unsigned PID                             : 13; // PID
	unsigned transport_scrambling_control    : 2;  // 加扰控制标志：表示TS流分组有效负载的加密模式。空包为‘00’
	unsigned adaptation_field_control        : 2;  // 适配域控制标志‘00’为ISO/IEC未来使用保留；
	unsigned continuity_counter              : 4;  // 连续性计数器
	unsigned adaptaion_field_length          : 8;  // 适配域长度
} TS_HEADER;

// pat 节目 
typedef struct TS_PAT_PROGRAM
{
	unsigned program_number                  : 16; // 节目号
	unsigned reserved                        : 3;  // 保留字段，固定为111
	unsigned PID                             : 13; // 节目号对应内容的PID值
} TS_PAT_PROGRAM;

// pat表
typedef struct TS_PAT
{
	unsigned table_id                        : 8;  // PAT表固定为0x00
	unsigned section_syntax_indicator        : 1;  // 段语法标志位，固定为1
	unsigned zero                            : 1;  // 固定为0
	unsigned reserved1                       : 2;  //  保留字段，固定为11
	unsigned section_length                  : 12; // 表示这个字节后面数据的长度,包括 CRC信息
	unsigned transport_stream_id             : 16; // 该传输流的ID
	unsigned reserved2                       : 2;  // 保留字段，固定为11
	unsigned version_number                  : 5;  // 版本号，固定为00000,有变化则版本号加1
	unsigned current_next_indicator          : 1;  // PAT是否有效,固定为1，表示这个PAT表可以用，如果为0则要等待下一个PAT表
	unsigned section_number                  : 8;  // 分段号码,最多256个分段
	unsigned last_section_number             : 8;  // 最后一个分段的号码
	unsigned networkPID                      : 16; // 网络PID
	unsigned CRC                             : 32; // CRC校验码
    int program_count                        : 8;  // 节目数量
	TS_PAT_PROGRAM *pPrograms;
} TS_PAT;

typedef struct TS_PMT_STREAM
{
	unsigned stream_type                     : 8;  // 流类型  h.264编码对应0x1b;aac编码对应0x0f;mp3编码对应0x03
	unsigned reserved1                       : 3;  // 保留字段，固定为111
	unsigned elementary_PID                  : 13; // 元素PID,与stream_type对应的PID
	unsigned reserved2                       : 4;  // 保留字段，固定为1111
	unsigned ES_info_length                  : 12; //  描述信息，指定为0x000表示没有
	FRAME_AV_TYPE av_type;
	unsigned char *pEsInfoBytes;
} TS_PMT_STREAM;

// pem表
typedef struct TS_PMT
{
	unsigned table_id                       : 8;  // PAT表固定为0x00
	unsigned section_syntax_indicator       : 1;  // 段语法标志位，固定为1
	unsigned zero                           : 1;  // 固定为0
	unsigned reserved1                      : 2;  //  保留字段，固定为11
	unsigned section_length                 : 12; // 表示这个字节后面数据的长度,包括 CRC信息
	unsigned program_number                 : 16; // 频道号码，表示当前的PMT关联到的频道，取值0x0001
	unsigned reserved2                      : 2;  // 保留字段，固定为11
	unsigned version_number                 : 5;  // 版本号，固定为00000，如果PAT有变化则版本号加1
	unsigned current_next_indicator         : 1;  // 是否有效
	unsigned section_number                 : 8;  // 分段号码
	unsigned last_section_number            : 8;  // 最后一个分段的号码
	unsigned reserved3                      : 3;  // 保留字段，固定为111
	unsigned PCR_PID                        : 13; // PCR(节目参考时钟)所在TS分组的PID，指定为视频PID
	unsigned reserved4                      : 4;  // 保留字段固定为 1111
	unsigned program_info_length            : 12; // 节目描述信息，指定为0x000表示没有
	unsigned char *pProgramInfoBytes;
	unsigned CRC                            : 32; // CRC校验码
	unsigned main_video_PID                 : 13; // 视频流PID
	unsigned main_audio_PID                 : 13; // 音频流PID
	int stream_count                        : 8;  // 流总数
	TS_PMT_STREAM *pStreams;                      // 流数据 
} TS_PMT;

typedef struct TS_PES_PACKET
{
	unsigned pes_start_code_prefix         :24; // 起始码，固定必须是'0000 0000 0000 0000 0000 0001' (0x000001)。用于标识包的开始。
	unsigned stream_id                     :8;  // 流ID
	unsigned PES_packet_length             :16; // PES包的长度
	unsigned twobit_10                     :2;  // 固定两位分割bit 0x2
	unsigned PES_scrambling_control        :2;  // 字段指示 PES 包有效载荷的加扰方式; PES 包头，其中包括任选字段只要存在，应不加扰。00 不加扰
	unsigned PES_priority                  :1;  // 指示在此 PES 包中该有效载荷的优先级。
	unsigned data_alignment_indicator      :1;  // 数据校准标志
	unsigned copyright                     :1;  // 版权保护标志
	unsigned original_or_copy              :1;  // 是否为复制
	unsigned PTS_DTS_flags                 :2;  // PTS(presentation time stamp 显示时间标签),DTS(decoding time stamp 解码时间标签)标志位
	unsigned ESCR_flag                     :1;  // 置于‘1’时指示 PES 包头中 ESCR 基准字段和 ESCR 扩展字段均存在。
	unsigned ES_rate_flag                  :1;  // 置于‘1’时指示 PES 包头中 ES_rate 字段存在。
	unsigned DSM_trick_mode_flag           :1;  // 特技方式
	unsigned additional_copy_info_flag     :1;  // 置于‘1’时指示 additional_copy_info 存在。
	unsigned PES_CRC_flag                  :1;  // 置于‘1’时指示 PES 包中 CRC 字段存在。
	unsigned PES_extension_flag            :1;  // 置于‘1’时指示 PES 包头中扩展字段存在。置于‘0’时指示此字段不存在
	unsigned PES_header_data_length        :8;  // 指示在此PES包头中包含的由任选字段和任意填充字节所占据的字节总数。
	unsigned long long PTS                 :33; // PTS(presentation time stamp 显示时间标签)
	unsigned long long DTS                 :33; // DTS(decoding time stamp 解码时间标签)标志位
	unsigned long long ESCR_base           :33; // 基本流时钟参考
	unsigned ESCR_extension                :9;  // 基本流时钟参考
	unsigned ES_rate                       :22; // ES 速率（基本流速率）
	unsigned trick_mode_control            :3;  // 3 比特字段，指示适用于相关视频流的特技方式
	unsigned field_id                      :2;  // 2 比特字段，指示哪些字段应予显示
	unsigned intra_slice_refresh           :1;  // 1 比特标志，置于‘1’时指示此 PES 包中视频数据的编码截面间可能存在丢失宏块
	unsigned frequency_truncation          :2;  // 指示在此 PES 包中编码视频数据时曾经使用的受限系数集
	unsigned rep_cntrl                     :5;  // 指示交错图像中每个字段应予显示的次数，或者连续图像应予显示的次数
	unsigned additional_copy_info          :7;  // 此 7 比特字段包含与版权信息有关的专用数据
	unsigned previous_PES_packet_CRC       :16; // 包含产生解码器中 16 寄存器零输出的 CRC 值
	FRAME_AV_TYPE av_type;
	unsigned stream_type;
} TS_PES_PACKET;

// ts解封装器
typedef struct TS_DEMUXER
{
	TS_PAT *global_pat;                   // 全局表
	TS_PMT *global_pmt;                   // 全局映射表
	HASH_MAP *global_buffer_map;          // 缓存map
	int cur_program_num;                  // 当前节目号
	FRAME_DATA_POOL *audio_pool;                // 帧池
	FRAME_DATA_POOL *video_pool;                // 帧池
	PRIORITY_QUEUE *pkt_queue;            // pes包队列
	TS_PAT_PROGRAM *temp_programs;        // 临时节目信息存储
	TS_PMT_STREAM * temp_streams;         // 临时流信息存储
} TS_DEMUXER;

// 解封装模块创建
TS_DEMUXER *ts_demuxer_create(FRAME_DATA_POOL *audio_pool, FRAME_DATA_POOL *video_pool);

// 解封装
int demux_ts_pkt(TS_DEMUXER *d, unsigned char *pTsBuf);

// 按指定节目号解封装
int demux_ts_pkt_by_program_num(TS_DEMUXER *d, unsigned char *pTsBuf, int programNum);

// 拉取 pes 包
FRAME_DATA * poll_pes_pkt(TS_DEMUXER *d);
FRAME_DATA * poll_pes_pkt_by_type(TS_DEMUXER *d, unsigned stream_type);

// 清空队列
void pes_queue_clean(TS_DEMUXER *d);

// 队列是否为空
int is_pes_queue_empty(TS_DEMUXER *d);

// 摧毁解封装模块
void ts_demuxer_destroy(TS_DEMUXER *d);
static void _free_ts_pat_program(TS_PAT_PROGRAM *pProgram);
static void _free_ts_pat(TS_PAT *pPat);
static void _free_ts_pmt_stream(TS_PMT_STREAM *pStream);
static void _free_ts_pmt(TS_PMT *pPmt);
void _free_ts_pes_pkt(TS_PES_PACKET *pPesPkt);

// 读取ts包头
static int _read_ts_head(TS_DEMUXER *d, unsigned char *pTsBuf, TS_HEADER *pHeader);

// 读取适配域
static int _read_adaption_field(TS_DEMUXER *d, unsigned char *pTsBuf, TS_HEADER *pHeader);

// 读取有效载荷
static int _read_payload(TS_DEMUXER *d, unsigned char *pTsBuf, TS_HEADER *pHeader);

// 解析PAT
static int _read_ts_PAT(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader);

// 提交pat表
static int _ts_pat_submit(TS_DEMUXER *d, TS_PAT pat);

// 解析PMT
static int _read_ts_PMT(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader);

// 提交pmt表
static int _ts_pmt_submit(TS_DEMUXER *d, TS_PMT pat);

// 输入pes载荷
static int _receive_pes_payload(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader);

// 解析pes包
static int _read_pes(TS_DEMUXER *d, BYTE_LIST * pPesByteList, TS_PMT_STREAM s);