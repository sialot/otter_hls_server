#include "ts_demuxer.h"

// 每次缓存的最大pes包数
const int BUFFER_COUNT = 200;

// 解封装模块创建
TS_DEMUXER * ts_demuxer_create(FRAME_DATA_POOL *audio_pool, FRAME_DATA_POOL *video_pool)
{
	TS_DEMUXER *d = (TS_DEMUXER *)malloc(sizeof(TS_DEMUXER));
	if (d == NULL)
	{
		printf("create_ts_demuxer failed. \n");
		return NULL;
	}
	d->global_pat = NULL;
	d->global_pmt = NULL;
	d->global_buffer_map = NULL;
	d->pkt_queue = NULL;
	d->cur_program_num = -1;
	d->audio_pool = audio_pool;
	d->video_pool = video_pool;
	d->temp_programs = NULL;
	d->temp_streams = NULL;

	HASH_MAP *map = hash_map_create();
	if (map == NULL)
	{
		printf("create_ts_demuxer failed. global_buffer_map create failed \n");
		ts_demuxer_destroy(d);
		return NULL;
	}
	d->global_buffer_map = map;

	PRIORITY_QUEUE *pkt_queue = priority_queue_create(BUFFER_COUNT, 0);
	if (pkt_queue == NULL)
	{
		printf("[%s]ts_pkt_queue init failed!\n", __FUNCTION__);
		hash_map_destroy(d->global_buffer_map);
		ts_demuxer_destroy(d);
		return NULL;
	}
	d->pkt_queue = pkt_queue;

	//printf("ts_demuxer_create complete! \n");
	return d;
}

// 解封装
int demux_ts_pkt(TS_DEMUXER *d, unsigned char * pTsBuf)
{

	// 解析ts头
	TS_HEADER header = { sizeof(TS_HEADER),0 };
	if (_read_ts_head(d, pTsBuf, &header) != 0)
	{
		return -1;
	}

	// 获取适配域
	if (_read_adaption_field(d, pTsBuf, &header) != 0)
	{
		return -1;
	}

	// 解析净荷
	if (_read_payload(d, pTsBuf, &header) != 0)
	{
		return -1;
	}

	return 0;
}

// 按指定节目号解封装
int demux_ts_pkt_by_program_num(TS_DEMUXER *d, unsigned char * pTsBuf, int programNum)
{
	d->cur_program_num = programNum;
	demux_ts_pkt(d, pTsBuf);
	return 0;
}

// 拉取 pes 包
FRAME_DATA * poll_pes_pkt(TS_DEMUXER * d)
{
	FRAME_DATA *pkt = priority_queue_poll(d->pkt_queue);
	return pkt;
}

// 按 视频\音频 拉pes包
FRAME_DATA * poll_pes_pkt_by_type(TS_DEMUXER * d, unsigned stream_type)
{
	FRAME_DATA *pkt = priority_queue_poll_by_type(d->pkt_queue, stream_type);
	return pkt;
}

// 清空队列
void pes_queue_clean(TS_DEMUXER * d)
{
	priority_queue_clean(d->pkt_queue);
}

// 队列是否为空
int is_pes_queue_empty(TS_DEMUXER *d)
{
	return is_priority_queue_empty(d->pkt_queue);
}

// 摧毁解封装模块
void ts_demuxer_destroy(TS_DEMUXER * d)
{
	if (d == NULL)
	{
		return;
	}

	// PAT
	if (d->global_pat != NULL) {
		_free_ts_pat(d->global_pat);
	}

	// PMT
	if (d->global_pmt != NULL) {
		_free_ts_pmt(d->global_pmt);
	}

	// MAP
	if (d->global_buffer_map != NULL) {
		hash_map_destroy(d->global_buffer_map);
	}

	// queue
	if (d->pkt_queue != NULL) {
		priority_queue_destroy(d->pkt_queue);
	}

	// temp
	free(d->temp_programs);
	free(d->temp_streams);

	// self
	free(d);
}

static void _free_ts_pat_program(TS_PAT_PROGRAM *pProgram)
{
	if (pProgram == NULL)
		return;
	free(pProgram);
	return;
}

static void _free_ts_pat(TS_PAT *pPat)
{
	if (pPat == NULL)
		return;
	free(pPat->pPrograms);
	free(pPat);
	return;
}

static void _free_ts_pmt_stream(TS_PMT_STREAM *pStream)
{
	if (pStream == NULL)
		return;
	if (pStream->pEsInfoBytes != NULL)
	{
		free(pStream->pEsInfoBytes);
	}
	free(pStream);
	return;
}

static void _free_ts_pmt(TS_PMT *pPmt)
{
	if (pPmt == NULL)
		return;

	if (pPmt->pProgramInfoBytes != NULL)
	{
		free(pPmt->pProgramInfoBytes);
	}
	_free_ts_pmt_stream(pPmt->pStreams);
	return;
}
void _free_ts_pes_pkt(TS_PES_PACKET *pPesPkt) {
	if (pPesPkt == NULL)
		return;
	free(pPesPkt);
	return;
}

// 读取ts包头
static int _read_ts_head(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader)
{
	pHeader->sync_byte = pTsBuf[0];
	if (pHeader->sync_byte != 0x47)
	{
		printf("sync_byte != 0x47\n");
		return -1;
	}

	pHeader->transport_error_indicator = pTsBuf[1] >> 7;
	pHeader->payload_unit_start_indicator = pTsBuf[1] >> 6 & 0x01;
	pHeader->transport_priority = pTsBuf[1] >> 5 & 0x01;
	pHeader->PID = (pTsBuf[1] & 0x1f) << 8 | pTsBuf[2];
	pHeader->transport_scrambling_control = pTsBuf[3] >> 6;
	pHeader->adaptation_field_control = pTsBuf[3] >> 4 & 0x03;
	pHeader->continuity_counter = pTsBuf[3] & 0x0f;
	return 0;
}

// 读取适配域
static int _read_adaption_field(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader)
{
	if (pHeader->adaptation_field_control == 0x2 ||
		pHeader->adaptation_field_control == 0x3)
	{
		pHeader->adaptaion_field_length = pTsBuf[4];
	}

	//  TODO 具体解析未实现
	//  ...A
	//  ...

	return 0;
}
// 读取有效载荷
static int _read_payload(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader)
{
	int rs = -1;

	// 看是否为PAT信息
	if (pHeader->PID == 0x0)
	{
		rs = _read_ts_PAT(d, pTsBuf, pHeader);
	}

	// 是否为 BAT/SDT 信息
	if (pHeader->PID == 0x11)
	{
		//  TODO 具体解析未实现
		//  ...
		//  ...
	}

	//  看是否为PMT信息
	if (d->global_pat != NULL)
	{
		for (int i = 0; i < d->global_pat->program_count; i++) {
			if (d->global_pat->pPrograms[i].PID == pHeader->PID) {
				rs = _read_ts_PMT(d, pTsBuf, pHeader);
				break;
			}
		}
	}

	// 看是否为PES信息
	if (d->global_pmt != NULL)
	{
		// 只解析主视频 和 主音频
		if (pHeader->PID == d->global_pmt->main_video_PID ||
			pHeader->PID == d->global_pmt->main_audio_PID)
		{
			rs = _receive_pes_payload(d, pTsBuf, pHeader);
		}
	}

	return rs;
}

// 解析PAT
static int _read_ts_PAT(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader)
{
	int start = 4;
	int adpataion_field_length = pTsBuf[4];

	// 同时又负载和适配域
	if (pHeader->adaptation_field_control == 0x3)
	{
		start = start + 1 + pHeader->adaptation_field_control;
	}

	// 有效载荷开始的位置应再偏移1+[length]个字节。
	if (pHeader->payload_unit_start_indicator == 0x01)
	{
		start = start + 1 + pTsBuf[start];
	}

	unsigned char *payload = pTsBuf + start;

	TS_PAT tempPat;
	tempPat.table_id = payload[0];
	tempPat.section_syntax_indicator = payload[1] >> 7 & 0x1;
	tempPat.zero = payload[1] >> 6 & 0x1;
	tempPat.reserved1 = payload[1] >> 4 & 0x3;
	tempPat.section_length = ((payload[1] & 0x0f) << 8) | payload[2];
	tempPat.transport_stream_id = (payload[3] << 8) | payload[4];
	tempPat.reserved2 = payload[5] >> 6 & 0x3;
	tempPat.version_number = payload[5] >> 1 & 0x1f;
	tempPat.current_next_indicator = payload[5] & 0x1;
	tempPat.section_number = payload[6];
	tempPat.last_section_number = payload[7];
	tempPat.pPrograms = NULL;

	// current_next_indicator 当前包无效
	if (tempPat.current_next_indicator != 0x1)
	{
		return 0;
	}

	if (tempPat.section_syntax_indicator != 0x1)
	{
		printf("不支持 section_syntax_indicator 为 0 \n");
		return -1;
	}

	// 有效负载总长度
	int len = 3 + tempPat.section_length;
	tempPat.CRC = payload[len - 4] << 24 | payload[len - 3] << 16 |	payload[len - 2] << 8 |	(payload[len - 1] & 0xFF);

	// 检测固定位
	if (tempPat.table_id != 0x00)
	{
		printf("pat parse error!table_id \n");
		return -1;
	}
	if (tempPat.zero != 0x0)
	{
		printf("pat parse error!zero \n");
		return -1;
	}

	// 提取循环部分字节数组
	int loopStartPos = 8;
	int loopLength = tempPat.section_length - 9;

	// 循环数据数组起始
	unsigned char * pLoopData = payload + loopStartPos;

	BYTE_LIST * pat_loop_data_buffer = (BYTE_LIST *)hash_map_get(d->global_buffer_map, pHeader->PID);

	// 当前段不是第一个pat包，加载旧字节数组
	if (tempPat.section_number != 0x00)
	{
		if (pat_loop_data_buffer != NULL) // buffer未初始化
		{
			byte_list_add_list(pat_loop_data_buffer, pLoopData, loopLength);
		}
	}
	else
	{
		if (pat_loop_data_buffer != NULL)
		{
			// 抛弃旧的buffer数据
			byte_list_clean(pat_loop_data_buffer);
		}
		else 
		{
			pat_loop_data_buffer = byte_list_create(loopLength);
			byte_list_add_list(pat_loop_data_buffer, pLoopData, loopLength);
			hash_map_put(d->global_buffer_map, pHeader->PID, pat_loop_data_buffer);
		}
	}

	// 当前为最后分段，解析循环数据
	if (tempPat.section_number == tempPat.last_section_number)
	{

		// 不为4的整数倍，数据有错误
		if ((pat_loop_data_buffer->used_len % 4) != 0)
		{
			printf("pat parse error!pat.loopData.length \n");
			return -1;
		}

		tempPat.program_count = pat_loop_data_buffer->used_len / 4;

		if (d->temp_programs == NULL)
		{
			TS_PAT_PROGRAM *temp_programs = (TS_PAT_PROGRAM *)malloc(sizeof(TS_PAT_PROGRAM) * tempPat.program_count);
			d->temp_programs = temp_programs;
		}

		tempPat.pPrograms = d->temp_programs;

		for (int i = 0; i < pat_loop_data_buffer->used_len; i += 4)
		{

			int program_number = ((pat_loop_data_buffer->pBytes[i] & 0xff) << 8) | (pat_loop_data_buffer->pBytes[i + 1] & 0xff);

			// 0x00 是NIT
			if (program_number == 0x00)
			{
				tempPat.networkPID = ((pat_loop_data_buffer->pBytes[i + 2] & 0x1f) << 8) | (pat_loop_data_buffer->pBytes[i + 3] & 0xff);
			}
			else
			{
				TS_PAT_PROGRAM prg;
				prg.program_number = program_number;
				prg.reserved = pat_loop_data_buffer->pBytes[i + 2] >> 5 & 0x3;
				prg.PID = ((pat_loop_data_buffer->pBytes[i + 2] & 0x1f) << 8) | (pat_loop_data_buffer->pBytes[i + 3] & 0xff);
				tempPat.pPrograms[i / 4] = prg;
			}
		}
			   
		// 清空data
		byte_list_clean(pat_loop_data_buffer);
	}

	_ts_pat_submit(d, tempPat);

	if ((d->cur_program_num == -1) && (d->global_pat->program_count >= 1)) {
		d->cur_program_num = d->global_pat->pPrograms[0].program_number;
	}

	return 0;
}

// 提交pat表
static int _ts_pat_submit(TS_DEMUXER *d, TS_PAT pat)
{
	if (d->global_pat == NULL)
	{
		d->global_pat = (TS_PAT *)malloc(sizeof(TS_PAT));
		if (d->global_pat == NULL)
		{
			return -1;
		}
		d->global_pat->pPrograms = NULL;
		d->global_pat->program_count = 0;
	}

	d->global_pat->table_id = pat.table_id;
	d->global_pat->section_syntax_indicator = pat.section_syntax_indicator;
	d->global_pat->zero = pat.zero;
	d->global_pat->reserved1 = pat.reserved1;
	d->global_pat->section_length = pat.section_length;
	d->global_pat->transport_stream_id = pat.transport_stream_id;
	d->global_pat->reserved2 = pat.reserved2;
	d->global_pat->version_number = pat.version_number;
	d->global_pat->current_next_indicator = pat.current_next_indicator;
	d->global_pat->section_number = pat.section_number;
	d->global_pat->last_section_number = pat.last_section_number;
	d->global_pat->networkPID = pat.networkPID;
	d->global_pat->CRC = pat.CRC;

	// 新节目更多
	if (d->global_pat->program_count < pat.program_count)
	{	
		if (d->global_pat->pPrograms != NULL)
		{
			TS_PAT_PROGRAM * pProgramsNew = realloc(d->global_pat->pPrograms, sizeof(TS_PAT_PROGRAM) * pat.program_count);
			if (pProgramsNew == NULL)
			{
				_free_ts_pat(d->global_pat);
				d->global_pat = NULL;
				return -1;
			}
			d->global_pat->pPrograms = pProgramsNew;
		}
		else
		{
			TS_PAT_PROGRAM * pProgramsNew = malloc(sizeof(TS_PAT_PROGRAM) * pat.program_count);
			if (pProgramsNew == NULL)
			{
				_free_ts_pat(d->global_pat);
				d->global_pat = NULL;
				return -1;
			}
			d->global_pat->pPrograms = pProgramsNew;
		}		
	}
	d->global_pat->program_count = pat.program_count;

	memcpy(d->global_pat->pPrograms, pat.pPrograms, sizeof(TS_PAT_PROGRAM) * pat.program_count);
	return 0;
}

// 解析PMT
int _read_ts_PMT(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader)
{
	int start = 4;
	int adpataion_field_length = pTsBuf[4];

	// 同时又负载和适配域
	if (pHeader->adaptation_field_control == 0x3)
	{
		start = start + 1 + pHeader->adaptation_field_control;
	}

	// 有效载荷开始的位置应再偏移1+[length]个字节。
	if (pHeader->payload_unit_start_indicator == 0x01)
	{
		start = start + 1 + pTsBuf[start];
	}

	unsigned char *payload = pTsBuf + start;

	TS_PMT tempPmt;
	tempPmt.table_id = payload[0];
	tempPmt.section_syntax_indicator = payload[1] >> 7 & 0x1;
	tempPmt.zero = payload[1] >> 6 & 0x1;
	tempPmt.reserved1 = payload[1] >> 4 & 0x3;
	tempPmt.section_length = ((payload[1] & 0x0f) << 8) | payload[2];
	tempPmt.program_number = (payload[3] << 8) | payload[4];
	tempPmt.reserved2 = payload[5] >> 6 & 0x3;
	tempPmt.version_number = payload[5] >> 1 & 0x1f;
	tempPmt.current_next_indicator = payload[5] & 0x1;
	tempPmt.section_number = payload[6];
	tempPmt.last_section_number = payload[7];
	tempPmt.reserved3 = payload[8] >> 5 & 0x7;
	tempPmt.PCR_PID = ((payload[8] & 0x1f) << 8) | payload[9];
	tempPmt.reserved4 = payload[10] >> 4 & 0xf;
	tempPmt.program_info_length = ((payload[10] & 0xf) << 8) | payload[11];
	tempPmt.pStreams = NULL;
	tempPmt.pProgramInfoBytes = NULL;

	// 不是当前节目的pmt包，舍弃
	if (d->cur_program_num != tempPmt.program_number)
	{
		return 0;
	}

	// current_next_indicator 当前包无效
	if (tempPmt.current_next_indicator != 0x1)
	{
		return 0;
	}

	if (tempPmt.section_syntax_indicator != 0x1) {
		printf("不支持 section_syntax_indicator 为 0 \n");
		return -1;
	}

	// 有效负载总长度
	int len = 3 + tempPmt.section_length;
	tempPmt.CRC = payload[len - 4] << 24 | payload[len - 3] << 16 |	payload[len - 2] << 8 |	(payload[len - 1] & 0xFF);

	// 检测固定位
	if (tempPmt.section_syntax_indicator != 0x1)
	{
		printf("pat parse error!section_syntax_indicator \n");
		return -1;
	}
	if (tempPmt.zero != 0x0)
	{
		printf("pat parse error!zero \n");
		return -1;
	}

	// programInfo提取
	if (tempPmt.program_info_length != 0x0)
	{
		// TODO
	}
	
	// 提取循环部分字节数组
	int loopStartPos = 12 + tempPmt.program_info_length;
	int loopLength = tempPmt.section_length - 13 - tempPmt.program_info_length;

	// 循环数据数组起始
	unsigned char * pLoopData = payload + loopStartPos;
	
	BYTE_LIST *pmt_loop_data_buffer = (BYTE_LIST *)hash_map_get(d->global_buffer_map, pHeader->PID);

	// 当前段不是第一个pat包，加载旧字节数组
	if (tempPmt.section_number != 0x00)
	{
		if (pmt_loop_data_buffer != NULL)  // buffer未初始化
		{
			byte_list_add_list(pmt_loop_data_buffer, pLoopData, loopLength);
		}
	}
	else
	{
		if (pmt_loop_data_buffer != NULL)   // 抛弃旧的buffer数据
		{
			byte_list_clean(pmt_loop_data_buffer);
		}
		else
		{
			pmt_loop_data_buffer = byte_list_create(loopLength);
			byte_list_add_list(pmt_loop_data_buffer, pLoopData, loopLength);
			hash_map_put(d->global_buffer_map, pHeader->PID, pmt_loop_data_buffer);
		}
	}

	// 当前为最后分段，解析循环数据
	if (tempPmt.section_number == tempPmt.last_section_number)
	{
		int pos = 0;
		int streamcount = 0;

		// 获取有多少流
		for (; pos < pmt_loop_data_buffer->used_len;)
		{
			int es_info_length = ((pmt_loop_data_buffer->pBytes[pos + 3] & 0xf) << 8) | pmt_loop_data_buffer->pBytes[pos + 4];
			if (es_info_length > 0) {
				pos += es_info_length;
			}
			pos += 5;
			streamcount++;
		}

		tempPmt.stream_count = streamcount;

		if (d->temp_streams == NULL)
		{
			TS_PMT_STREAM * temp_streams = (TS_PMT_STREAM *)malloc(sizeof(TS_PMT_STREAM) * tempPmt.stream_count);
			if (temp_streams == NULL)
			{
				return -1;
			}
			d->temp_streams = temp_streams;
		}

		tempPmt.pStreams = d->temp_streams;

		pos = 0;
		streamcount = 0;
		for (; pos < pmt_loop_data_buffer->used_len;)
		{

			TS_PMT_STREAM s;
			s.stream_type = pmt_loop_data_buffer->pBytes[pos];
			s.reserved1 = (pmt_loop_data_buffer->pBytes[pos + 1] >> 5) & 0x7;
			s.elementary_PID = ((pmt_loop_data_buffer->pBytes[pos + 1] & 0x1f) << 8) | pmt_loop_data_buffer->pBytes[pos + 2];
			s.reserved2 = (pmt_loop_data_buffer->pBytes[pos + 3] >> 4) & 0xf;
			s.ES_info_length = ((pmt_loop_data_buffer->pBytes[pos + 3] & 0xf) << 8) | pmt_loop_data_buffer->pBytes[pos + 4];
			s.pEsInfoBytes = NULL;	// TODO 暂未解析

			if (s.stream_type == 0x1b)
			{
				s.av_type = VIDEO;
			}
			if (s.stream_type == 0x0f)
			{
				s.av_type = AUDIO;
			}

			if (s.ES_info_length > 0) {

				// TODO 暂未解析
				pos += s.ES_info_length;
			}
			tempPmt.pStreams[streamcount] = s;
			pos += 5;
			streamcount++;
		}
		
		// 清空data
		byte_list_clean(pmt_loop_data_buffer);
	}
	_ts_pmt_submit(d, tempPmt);

	/* TEST LOG 
	for (int i = 0; i < d->global_pmt->stream_count; i++) {

		printf("stream_type:%d,elementary_PID:%d,ES_info_length:%d\n",
			d->global_pmt->pStreams[i].stream_type,
			d->global_pmt->pStreams[i].elementary_PID,
			d->global_pmt->pStreams[i].ES_info_length);
	}
	*/
	return 0;
}

// 提交pmt表
int _ts_pmt_submit(TS_DEMUXER *d, TS_PMT pmt)
{
	if (d->global_pmt == NULL)
	{
		d->global_pmt = (TS_PMT *)malloc(sizeof(TS_PMT));
		if (d->global_pmt == NULL)
		{
			return -1;
		}
		d->global_pmt->pStreams = NULL;
		d->global_pmt->pProgramInfoBytes = NULL;
		d->global_pmt->stream_count = 0;
	}

	d->global_pmt->table_id = pmt.table_id;
	d->global_pmt->section_syntax_indicator = pmt.section_syntax_indicator;
	d->global_pmt->zero = pmt.zero;
	d->global_pmt->reserved1 = pmt.reserved1;
	d->global_pmt->section_length = pmt.section_length;
	d->global_pmt->program_number = pmt.program_number;
	d->global_pmt->reserved2 = pmt.reserved2;
	d->global_pmt->version_number = pmt.version_number;
	d->global_pmt->current_next_indicator = pmt.current_next_indicator;
	d->global_pmt->section_number = pmt.section_number;
	d->global_pmt->last_section_number = pmt.last_section_number;
	d->global_pmt->reserved3 = pmt.reserved3;
	d->global_pmt->PCR_PID = pmt.PCR_PID;
	d->global_pmt->reserved4 = pmt.reserved4;
	d->global_pmt->program_info_length = pmt.program_info_length;

	// 新的流变多
	if (d->global_pmt->stream_count < pmt.stream_count)
	{
		if (d->global_pmt->pStreams != NULL)
		{
			TS_PMT_STREAM * pStreamsNew = realloc(d->global_pmt->pStreams, sizeof(TS_PMT_STREAM) * pmt.stream_count);
			if (pStreamsNew == NULL)
			{
				_free_ts_pmt(d->global_pmt);
				d->global_pmt = NULL;
				return -1;
			}
			d->global_pmt->pStreams = pStreamsNew;
		}
		else
		{
			TS_PMT_STREAM * pStreamsNew = malloc(sizeof(TS_PMT_STREAM) * pmt.stream_count);
			if (pStreamsNew == NULL)
			{
				_free_ts_pmt(d->global_pmt);
				d->global_pmt = NULL;
				return -1;
			}
			d->global_pmt->pStreams = pStreamsNew;
		}
	}
	d->global_pmt->stream_count = pmt.stream_count;

	memcpy(d->global_pmt->pStreams, pmt.pStreams, sizeof(TS_PMT_STREAM) * pmt.stream_count);

	int video_found = 0;
	int audio_found = 0;

	// 设置视频、音频
	for(int i = 0; i<  d->global_pmt->stream_count; i++)
	{ 
		// h.264编码对应0x1b
		// aac编码对应0x0f
		if (d->global_pmt->pStreams[i].stream_type == 0x1b && !video_found)
		{
			d->global_pmt->main_video_PID = d->global_pmt->pStreams[i].elementary_PID;
			video_found = 1;
		}
		if (d->global_pmt->pStreams[i].stream_type == 0x0f && !audio_found)
		{
			d->global_pmt->main_audio_PID = d->global_pmt->pStreams[i].elementary_PID;
			audio_found = 1;
		}
	}

	return 0;
}

// 输入pes载荷
int _receive_pes_payload(TS_DEMUXER *d, unsigned char * pTsBuf, TS_HEADER * pHeader)
{
	int start = 4;
	int adpataion_field_length = pTsBuf[4];
	if (pHeader->adaptation_field_control == 0x3)
	{
		start = start + 1 + pHeader->adaptaion_field_length;
	}
	unsigned char *payload = pTsBuf + start;
	int payload_len = 188 - start;
	
	// ts包中含有新pes包头
	if (pHeader->payload_unit_start_indicator == 0x1) {

		unsigned PES_packet_length = payload[4] << 8 | payload[5] ;
		unsigned finish_len = 0;
		if (PES_packet_length != 0)
		{
			finish_len = PES_packet_length + 6;
		}

		// 取出旧pes包缓存数据,并处理清空
		BYTE_LIST *pesBuffer = (BYTE_LIST *)hash_map_get(d->global_buffer_map, pHeader->PID);

		if (pesBuffer != NULL) {

			if (pesBuffer->used_len > 0)
			{
				if (pHeader->PID == d->global_pmt->main_audio_PID 
					|| pHeader->PID == d->global_pmt->main_video_PID)
				{
					for (int i = 0; i < d->global_pmt->stream_count; i++)
					{
						if (pHeader->PID == d->global_pmt->pStreams[i].elementary_PID)
						{
							_read_pes(d, pesBuffer, d->global_pmt->pStreams[i]);
						}
					}
				}
			}

			// 释放空间
			byte_list_clean(pesBuffer);
		}
		else
		{
			// 创建缓存
			pesBuffer = byte_list_create(finish_len);
			hash_map_put(d->global_buffer_map, pHeader->PID, pesBuffer);
		}
		
		pesBuffer->finish_len = finish_len;

		// 缓存中追加新数据
		byte_list_add_list(pesBuffer, payload, payload_len);

		// 调试信息打印
		int pesStartPrefix = (payload[0] << 16) | (payload[1] << 8) | (payload[2]);
		int stream_id = payload[3];
		//printf("%d (%d) > pre: %#X  stream_id: %d  PES_packet_length:%d  data: %d \n",	pHeader->PID, pHeader->payload_unit_start_indicator, pesStartPrefix, stream_id, PES_packet_length, payload_len);
	}
	else
	{

		// 取出旧pes包缓存数据
		BYTE_LIST *pesBuffer = (BYTE_LIST *)hash_map_get(d->global_buffer_map, pHeader->PID);
		if (pesBuffer != NULL)
		{
			// 缓存中追加新数据
			byte_list_add_list(pesBuffer, payload, payload_len);

			/*System.out.println(p.PID +"(" + p.payload_unit_start_indicator + ")" + "> data:" + + tsPayload.length + " loading:" +
					olderBuffer.buffer.length + "/" + (olderBuffer.PES_packet_length + 6));*/

			if (is_byte_list_finish(pesBuffer)) {

				if (pHeader->PID == d->global_pmt->main_audio_PID
					|| pHeader->PID == d->global_pmt->main_video_PID)
				{
					for (int i = 0; i < d->global_pmt->stream_count; i++)
					{
						if (pHeader->PID == d->global_pmt->pStreams[i].elementary_PID)
						{
							_read_pes(d, pesBuffer, d->global_pmt->pStreams[i]);
						}
					}
				}

				// 释放空间
				byte_list_clean(pesBuffer);
			}
		}
		else {
			printf("_receive_pes_payload:can't append data. \n");
			return -1;
		}
	}

	return 0;
}

// 解析pes包
int _read_pes(TS_DEMUXER *d, BYTE_LIST * pPesByteList, TS_PMT_STREAM s)
{
	TS_PES_PACKET tp;
	unsigned char * pl = pPesByteList->pBytes;

	tp.pes_start_code_prefix = pl[0] << 16 | pl[1] << 8	| pl[2];
	if (tp.pes_start_code_prefix != 0x001) {
		return -1;
	}
	tp.stream_id = pl[3];
	tp.PES_packet_length = pl[4] << 8 | pl[5];
	tp.twobit_10 = pl[6] >> 6 & 0x3;
	tp.PES_scrambling_control = pl[6] >> 4 & 0x3;
	tp.PES_priority = pl[6] >> 3 & 0x1;
	tp.data_alignment_indicator = pl[6] >> 2 & 0x1;
	tp.copyright = pl[6] >> 1 & 0x1;
	tp.original_or_copy = pl[6] & 0x1;
	tp.PTS_DTS_flags = pl[7] >> 6 & 0x3;
	tp.ESCR_flag = pl[7] >> 5 & 0x1;
	tp.ES_rate_flag = pl[7] >> 4 & 0x1;
	tp.DSM_trick_mode_flag = pl[7] >> 3 & 0x1;
	tp.additional_copy_info_flag = pl[7] >> 2 & 0x1;
	tp.PES_CRC_flag = pl[7] >> 1 & 0x1;
	tp.PES_extension_flag = pl[7] & 0x1;
	tp.PES_header_data_length = pl[8];
	tp.av_type = s.av_type;
	tp.stream_type = s.stream_type;
	
	// 可选域字节索引
	int opt_field_idx = 9;

	// PTS(presentation time stamp 显示时间标签)
	// DTS(decoding time stamp 解码时间标签)标志位
	if (tp.PTS_DTS_flags == 0x2) {
		tp.PTS = (pl[opt_field_idx] >> 1 & 0x7) << 30 |
			pl[opt_field_idx + 1] << 22 |
			(pl[opt_field_idx + 2] >> 1 & 0x7f) << 15 |
			pl[opt_field_idx + 3] << 7 |
			(pl[opt_field_idx + 4] >> 1 & 0x7f);
		opt_field_idx += 5;
		tp.DTS = tp.PTS;
	}
	else if (tp.PTS_DTS_flags == 0x3) {
		tp.PTS = (pl[opt_field_idx] >> 1 & 0x7) << 30 |
			pl[opt_field_idx + 1] << 22 |
			(pl[opt_field_idx + 2] >> 1 & 0x7f) << 15 |
			pl[opt_field_idx + 3] << 7 |
			(pl[opt_field_idx + 4] >> 1 & 0x7f);

		tp.DTS = (pl[opt_field_idx + 5] >> 1 & 0x7) << 30 |
			pl[opt_field_idx + 6] << 22 |
			(pl[opt_field_idx + 7] >> 1 & 0x7f) << 15 |
			pl[opt_field_idx + 8] << 7 |
			(pl[opt_field_idx + 9] >> 1 & 0x7f);
		opt_field_idx += 10;
	}

	// 基本流时钟参考
	// 00111011 11111111 11111011 11111111 11111011 11111110
	// 111 11 11111111 11111 11 11111111 11111
	// 11 1111111
	if (tp.ESCR_flag == 0x1) {
		tp.ESCR_base = (pl[opt_field_idx] >> 3 & 0x7) << 30 |
			(pl[opt_field_idx] & 0x3) << 28 |
			pl[opt_field_idx + 1] << 20 |
			(pl[opt_field_idx + 2] >> 3 & 0x1f) << 15 |
			(pl[opt_field_idx + 2] & 0x3) << 13 |
			pl[opt_field_idx + 3] << 5 |
			(pl[opt_field_idx + 4] >> 3 & 0x1f);
		tp.ESCR_extension = (pl[opt_field_idx + 4] & 0x3) << 7 |
			(pl[opt_field_idx + 5] >> 7 & 0x7f);

		opt_field_idx += 6;
	}

	// ES 速率（基本流速率）
	// 01111111 11111111 11111110
	if (tp.ES_rate_flag == 0x1) {
		tp.ES_rate = (pl[opt_field_idx] & 0x7f) << 15 |
			pl[opt_field_idx + 1] << 7 |
			(pl[opt_field_idx + 2] >> 1 & 0x7f);

		opt_field_idx += 3;
	}

	// 相关视频流的特技方式
	if (tp.DSM_trick_mode_flag == 0x1) {

		// '000' 快进
		// '001' 慢动作
		// '010' 冻结帧
		// '011' 快速反向
		// '100' 慢反向
		tp.trick_mode_control = pl[opt_field_idx] >> 5 & 0x7;
		if (tp.trick_mode_control == 0x0) {
			tp.field_id = pl[opt_field_idx] >> 3 & 0x3;
			tp.intra_slice_refresh = pl[opt_field_idx] >> 2 & 0x1;
			tp.frequency_truncation = pl[opt_field_idx] & 0x3;
		}
		else if (tp.trick_mode_control == 0x1) {
			tp.rep_cntrl = pl[opt_field_idx] & 0x1f;
		}
		else if (tp.trick_mode_control == 0x2) {
			tp.field_id = pl[opt_field_idx] >> 3 & 0x3;
		}
		else if (tp.trick_mode_control == 0x3) {
			tp.field_id = pl[opt_field_idx] >> 3 & 0x3;
			tp.intra_slice_refresh = pl[opt_field_idx] >> 2 & 0x1;
			tp.frequency_truncation = pl[opt_field_idx] & 0x3;
		}
		else if (tp.trick_mode_control == 0x4) {
			tp.rep_cntrl = pl[opt_field_idx] & 0x1f;
		}
		opt_field_idx += 1;
	}

	if (tp.additional_copy_info_flag == 0x1) {
		tp.additional_copy_info = pl[opt_field_idx] & 0x7f;
		opt_field_idx += 1;
	}

	if (tp.PES_CRC_flag == 0x1) {
		tp.previous_PES_packet_CRC = pl[opt_field_idx] << 8 |
			pl[opt_field_idx];
	}

	if (tp.PES_extension_flag == 0x1) {
		return -1;
	}

	// 截取payload
	int dataBegin = 9 + tp.PES_header_data_length;
	int es_data_len = pPesByteList->used_len - dataBegin;


	FRAME_DATA *fdata = NULL;

	if (tp.av_type == AUDIO)
	{
		fdata = frame_data_pool_borrow(d->audio_pool);
	}
	else
	{
		fdata = frame_data_pool_borrow(d->video_pool);
	}

	if (fdata == NULL)
	{
		return -1;
	}
	frame_data_set(fdata, tp.av_type, tp.stream_type, tp.DTS, tp.PTS, pPesByteList->pBytes + dataBegin, es_data_len);

	int add_res = priority_queue_push(d->pkt_queue, fdata, (fdata->dts / 90));
	if (add_res == -1)
	{
		printf("push pes failed!\n");
		return -1;
	}
	return 0;
}