package ts

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	errors "../errors"
)

// TS包字节数
const TS_PKG_SIZE int = 188
const TS_RELOAD_NUM = 30000

// 错误码
const ERROR_CODE_TS_HEADER_READ_FAILED int = 0
const ERROR_CODE_TS_PAT_READ_FAILED = 1
const ERROR_CODE_TS_PMT_READ_FAILED = 2
const ERROR_CODE_TS_PES_READ_FAILED = 3

// ts 头
type TsHeader struct {
	sync_byte                    uint8  //8 同步字节：固定为0x47;
	transport_error_indicator    uint8  //1 传输错误标志：‘1’表示在相关的传输包中至少有一个不可纠正的错误位。
	payload_unit_start_indicator uint8  //1 负载起始标志：在前4个字节之后会有一个调整字节，其的数值为后面调整字段的长度length。
	transport_priority           uint8  //1 传输优先级标志
	PID                          uint16 //13 PID
	transport_scrambling_control uint8  //2 加扰控制标志：表示TS流分组有效负载的加密模式。空包为‘00’
	adaptation_field_control     uint8  //2 适配域控制标志‘00’为ISO/IEC未来使用保留；
	continuity_counter           uint8  //4 连续性计数器
	adaptaion_field_length       uint8  //8 适配域长度
}

// PAT 中的 Program
type TsPatProgram struct {
	program_number uint16 //16 节目号
	reserved       uint8  //3 保留字段，固定为111
	PID            uint16 //16 节目号对应内容的PID值
}

// PAT 表
type TsPat struct {
	table_id                 uint8  //8 PAT表固定为0x00
	section_syntax_indicator uint8  //1 段语法标志位，固定为1
	zero                     uint8  //1 固定为0
	reserved1                uint8  //2 保留字段，固定为11
	section_length           uint16 //12 表示这个字节后面数据的长度,包括 CRC信息
	transport_stream_id      uint16 //16 该传输流的ID
	reserved2                uint8  //2 保留字段，固定为11
	version_number           uint8  //5 版本号，固定为00000,有变化则版本号加1
	current_next_indicator   uint8  //1 PAT是否有效,固定为1，表示这个PAT表可以用，如果为0则要等待下一个PAT表
	section_number           uint8  //8 分段号码,最多256个分段
	last_section_number      uint8  //8 最后一个分段的号码
	networkPID               uint16 //16 网络PID
	CRC                      uint32 //32 CRC校验码
	program_count            uint8  //8 节目数量
	pLoopData                []byte //循环数据
	programs                 []TsPatProgram
}

// 视频流信息结构体
type TsStream struct {
	stream_type    uint8  //8 流类型  h.264编码对应0x1b;aac编码对应0x0f;mp3编码对应0x03
	reserved1      uint8  //3 保留字段，固定为111
	elementary_PID uint16 //13 元素PID,与stream_type对应的PID
	reserved2      uint8  //4 保留字段，固定为1111
	ES_info_length uint16 //12  描述信息，指定为0x000表示没有
}

// pmt表
type TsPmt struct {
	table_id                 uint8      //8 PAT表固定为0x00
	section_syntax_indicator uint8      //1 段语法标志位，固定为1
	zero                     uint8      //1 固定为0
	reserved1                uint8      //2  保留字段，固定为11
	section_length           uint16     //12 表示这个字节后面数据的长度,包括 CRC信息
	program_number           uint16     //16 频道号码，表示当前的PMT关联到的频道，取值0x0001
	reserved2                uint8      //2 保留字段，固定为11
	version_number           uint8      //5 版本号，固定为00000，如果PAT有变化则版本号加1
	current_next_indicator   uint8      //1 是否有效
	section_number           uint8      //8 分段号码
	last_section_number      uint8      //8 最后一个分段的号码
	reserved3                uint8      //3 保留字段，固定为111
	PCR_PID                  uint16     //13 PCR(节目参考时钟)所在TS分组的PID，指定为视频PID
	reserved4                uint8      //4 保留字段固定为 1111
	program_info_length      uint16     //12 节目描述信息，指定为0x000表示没有
	CRC                      uint32     //32 CRC校验码
	stream_count             uint8      //8 流总数
	pLoopData                []byte     //循环数据
	streams                  []TsStream // 流数据
}

// pes数据结构体
type TsPes struct {
	pes_start_code_prefix     uint32 //24 起始码，固定必须是'0000 0000 0000 0000 0000 0001' (0x000001)。用于标识包的开始。
	stream_id                 uint8  //8 流ID
	PES_packet_length         uint16 //16 PES包的长度
	twobit_10                 uint8  //2 固定两位分割bit 0x2
	PES_scrambling_control    uint8  //2 字段指示 PES 包有效载荷的加扰方式; PES 包头，其中包括任选字段只要存在，应不加扰。00 不加扰
	PES_priority              uint8  //1 指示在此 PES 包中该有效载荷的优先级。
	data_alignment_indicator  uint8  //1 数据校准标志
	copyright                 uint8  //1 版权保护标志
	original_or_copy          uint8  //1 是否为复制
	PTS_DTS_flags             uint8  //2 PTS(presentation time stamp 显示时间标签),DTS(decoding time stamp 解码时间标签)标志位
	ESCR_flag                 uint8  //1 置于‘1’时指示 PES 包头中 ESCR 基准字段和 ESCR 扩展字段均存在。
	ES_rate_flag              uint8  //1 置于‘1’时指示 PES 包头中 ES_rate 字段存在。
	DSM_trick_mode_flag       uint8  //1 特技方式
	additional_copy_info_flag uint8  //1 置于‘1’时指示 additional_copy_info 存在。
	PES_CRC_flag              uint8  //1 置于‘1’时指示 PES 包中 CRC 字段存在。
	PES_extension_flag        uint8  //1 置于‘1’时指示 PES 包头中扩展字段存在。置于‘0’时指示此字段不存在
	PES_header_data_length    uint8  //8 指示在此PES包头中包含的由任选字段和任意填充字节所占据的字节总数。
	PTS                       uint64 //33 PTS(presentation time stamp 显示时间标签)
	DTS                       uint64 //33 DTS(decoding time stamp 解码时间标签)标志位
	ESCR_base                 uint64 //33 基本流时钟参考
	ESCR_extension            uint16 //9 基本流时钟参考
	ES_rate                   uint32 //22 ES 速率（基本流速率）
	trick_mode_control        uint8  //3 3 比特字段，指示适用于相关视频流的特技方式
	field_id                  uint8  //2 2 比特字段，指示哪些字段应予显示
	intra_slice_refresh       uint8  //1 1 比特标志，置于‘1’时指示此 PES 包中视频数据的编码截面间可能存在丢失宏块
	frequency_truncation      uint8  //2 指示在此 PES 包中编码视频数据时曾经使用的受限系数集
	rep_cntrl                 uint8  //5 指示交错图像中每个字段应予显示的次数，或者连续图像应予显示的次数
	additional_copy_info      uint8  //7 此 7 比特字段包含与版权信息有关的专用数据
	previous_PES_packet_CRC   uint16 //16 包含产生解码器中 16 寄存器零输出的 CRC 值
}

// TS解封装器
type TsDemuxer struct {
	globalPat     TsPat             // 全局PAT表
	globalPmt     TsPmt             // 全局PMT表
	bufferMap     map[uint16][]byte // 全局ts buffer临时存储，key PID,值 byte数据切片
	curPesLen     int               // 当前pes结束长度
	curProgramPID int
	curVideoPID   int
	curAudioPID   int
}

func (d *TsDemuxer) Init() {
	d.bufferMap = make(map[uint16][]byte)
	d.curPesLen = -1
	d.curProgramPID = -1
	d.curVideoPID = -1
	d.curAudioPID = -1
}

//处理文件
func (d *TsDemuxer) ProcessFile(file *os.File) {

	fmt.Printf("ts_demuxer.processFile start ! \n")

	// 预加载ts包字节 切片
	preLoadData := make([]byte, TS_PKG_SIZE*TS_RELOAD_NUM)

	// 取ts文件
	for {
		_, err := file.Read(preLoadData)
		//fmt.Printf("LoadData, size：%d \n", len(preLoadData))

		// 读取文件失败
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		// 解封装
		var i int
		for i = 0; i < TS_RELOAD_NUM; i++ {
			var pKgBuf []byte = preLoadData[i*188 : (i*188 + 188)]
			d.DemuxPkg(pKgBuf)
		}
	}

	fmt.Printf("ts_demuxer.processFile finish ! \n")
}

// 解封装
func (d *TsDemuxer) DemuxPkg(pKgBuf []byte) error {

	// 获取包头
	header, err := d.readTsHeader(pKgBuf)
	if err != nil {
		return err
	}

	// 获取适配域
	adpReadErr := d.readAdaptionField(pKgBuf, &header)
	if adpReadErr != nil {
		return adpReadErr
	}

	// 获取有效载荷, adaptation_field_control 01,11 代表有有效载荷
	if header.adaptation_field_control == 0x01 || header.adaptation_field_control == 0x03 {
		payloadReadErr := d.readPayload(pKgBuf, &header)
		if payloadReadErr != nil {
			return payloadReadErr
		}
	}

	return nil
}

// 解析TS包头
func (d *TsDemuxer) readTsHeader(pKgBuf []byte) (TsHeader, error) {
	var header TsHeader
	header.sync_byte = pKgBuf[0]

	// 不是有效的ts包，抛弃
	if header.sync_byte != 0x47 {
		err := errors.NewError(ERROR_CODE_TS_HEADER_READ_FAILED, "TsHeader read failed!")
		return header, err
	}

	header.transport_error_indicator = pKgBuf[1] >> 7
	header.payload_unit_start_indicator = pKgBuf[1] >> 6 & 0x01
	header.transport_priority = pKgBuf[1] >> 5 & 0x01
	header.PID = uint16(pKgBuf[1]&0x1f)<<8 | uint16(pKgBuf[2])
	header.transport_scrambling_control = pKgBuf[3] >> 6
	header.adaptation_field_control = pKgBuf[3] >> 4 & 0x03
	header.continuity_counter = pKgBuf[3] & 0x0f

	return header, nil
}

// 解析适配域
func (d *TsDemuxer) readAdaptionField(pKgBuf []byte, pHeader *TsHeader) error {

	if pHeader.adaptation_field_control == 0x2 || pHeader.adaptation_field_control == 0x3 {
		pHeader.adaptaion_field_length = pKgBuf[4]
	}

	//  TODO 具体解析未实现
	return nil
}

// 解析有效载荷
func (d *TsDemuxer) readPayload(pKgBuf []byte, pHeader *TsHeader) error {

	// 负载信息起始索引
	var start int = 4

	// 同时存在负载和适配域
	if pHeader.adaptation_field_control == 0x3 {
		start = start + 1 + int(pHeader.adaptaion_field_length)
	}

	// 看是否为PAT信息
	if pHeader.PID == 0x0 {

		// 对于PSI,payload_unit_start_indicator 为1时
		// 有效载荷开始的位置应再偏移 1 个字节,pointer_field。
		if pHeader.payload_unit_start_indicator == 0x01 {
			start = start + 1
		}

		d.readTsPAT(pKgBuf[start:len(pKgBuf)], pHeader)
	}

	// 是否为 BAT/SDT 信息
	if pHeader.PID == 0x11 {
		//  TODO 具体解析未实现
		//  ...
		//  ...
	}

	// 看是否为PMT信息
	if d.curProgramPID == int(pHeader.PID) {

		// 对于PSI,payload_unit_start_indicator 为1时
		// 有效载荷开始的位置应再偏移 1 个字节,pointer_field。
		if pHeader.payload_unit_start_indicator == 0x01 {
			start = start + 1
		}
		d.readTsPMT(pKgBuf[start:len(pKgBuf)], pHeader)
	}

	// 看是否为PES信息，只解析主视频 和 主音频
	if int(pHeader.PID) == d.curVideoPID || int(pHeader.PID) == d.curAudioPID {
		d.readPesPayload(pKgBuf[start:len(pKgBuf)], pHeader)
	}

	return nil
}

// 解析PAT表数据
func (d *TsDemuxer) readTsPAT(payload []byte, pHeader *TsHeader) error {

	// 获取临时PAT表
	table_id := payload[0]
	section_syntax_indicator := payload[1] >> 7 & 0x1
	zero := payload[1] >> 6 & 0x1
	reserved1 := payload[1] >> 4 & 0x3
	section_length := uint16(payload[1]&0x0f)<<8 | uint16(payload[2])
	transport_stream_id := uint16(payload[3])<<8 | uint16(payload[4])
	reserved2 := payload[5] >> 6 & 0x3
	version_number := payload[5] >> 1 & 0x1f
	current_next_indicator := payload[5] & 0x1
	section_number := payload[6]
	last_section_number := payload[7]
	var networkPID uint16
	var CRC uint32
	var program_count uint8

	// current_next_indicator 当前包无效
	// PAT是否有效,固定为1，表示这个PAT表可以用，如果为0则要等待下一个PAT表
	if current_next_indicator != 0x1 {
		return nil
	}

	// 检测三个固定位
	if table_id != 0x00 {
		err := errors.NewError(ERROR_CODE_TS_PAT_READ_FAILED, "pat parse error!table_id!")
		fmt.Printf("%s \n", err.ErrMsg)
		return err
	}
	if zero != 0x0 {
		err := errors.NewError(ERROR_CODE_TS_PAT_READ_FAILED, "pat parse error!zero!")
		fmt.Printf("%s \n", err.ErrMsg)
		return err
	}
	if section_syntax_indicator != 0x1 {
		err := errors.NewError(ERROR_CODE_TS_PAT_READ_FAILED, "不支持 section_syntax_indicator 为 0!")
		fmt.Printf("%s \n", err.ErrMsg)
		return err
	}

	// 有效负载总长度
	var plen int = 3 + int(section_length)
	CRC = uint32(payload[plen-4])<<24 | uint32(payload[plen-3])<<16 | uint32(payload[plen-2])<<8 | (uint32(payload[plen-1]) & 0xFF)

	// 提取循环部分字节数组
	var loopStartPos int = 8
	var loopLength int = int(section_length) - 9

	// 提取循环数据，开始为loopStartPos， 结束为 loopStartPos + loopLength
	var pLoopData []byte = payload[loopStartPos : loopStartPos+loopLength]

	// 追加TS 分段语法缓存
	d.storeTsSectionData(section_number, pLoopData, pHeader)

	// 当前为最后分段，解析循环数据
	if section_number == last_section_number {

		var loopDataBuffer []byte = d.bufferMap[pHeader.PID]

		// 校验循环数据是否有变化，pat数据没变化，取消解析
		if d.globalPat.pLoopData != nil && bytes.Equal(d.globalPat.pLoopData, loopDataBuffer) {

			// 清空缓存数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
			return nil
		}

		// 校验循环数据长度，如果不为4的整数倍，数据有错误
		if len(loopDataBuffer)%4 != 0 {

			err := errors.NewError(ERROR_CODE_TS_PAT_READ_FAILED, "pat parse error!pat.loopData.length!")
			fmt.Printf("%s \n", err.ErrMsg)
			return err
		}

		// 根据缓存计算program数
		program_count = uint8(len(loopDataBuffer) / 4)

		var programs []TsPatProgram = make([]TsPatProgram, program_count)

		var i int
		for i = 0; i < len(loopDataBuffer); i += 4 {

			var program_number uint16 = uint16(loopDataBuffer[i]&0xff)<<8 | uint16(loopDataBuffer[i+1]&0xff)

			// 0x00 是NIT
			if program_number == 0x00 {
				networkPID = uint16(loopDataBuffer[i+2]&0x1f)<<8 | uint16(loopDataBuffer[i+3]&0xff)
			} else {
				var prg TsPatProgram
				prg.program_number = program_number
				prg.reserved = loopDataBuffer[i+2] >> 5 & 0x3
				prg.PID = uint16(loopDataBuffer[i+2]&0x1f)<<8 | uint16(loopDataBuffer[i+3]&0xff)
				programs[i/4] = prg
			}
		}

		// 提交临时PAT表到全局PAT表
		d.globalPat.table_id = table_id
		d.globalPat.section_syntax_indicator = section_syntax_indicator
		d.globalPat.zero = zero
		d.globalPat.reserved1 = reserved1
		d.globalPat.section_length = section_length
		d.globalPat.transport_stream_id = transport_stream_id
		d.globalPat.reserved2 = reserved2
		d.globalPat.version_number = version_number
		d.globalPat.current_next_indicator = current_next_indicator
		d.globalPat.section_number = section_number
		d.globalPat.last_section_number = last_section_number
		d.globalPat.networkPID = networkPID
		d.globalPat.CRC = CRC
		d.globalPat.program_count = program_count
		d.globalPat.pLoopData = loopDataBuffer
		d.globalPat.programs = programs

		// 获取第一个节目作为解析目标，只解析第一个节目
		d.curProgramPID = int(d.globalPat.programs[0].PID)

		fmt.Printf("识别到PAT表，PID：%d \n", pHeader.PID)
		fmt.Printf("识别到当前Program，PID: %d \n", d.curProgramPID)
		fmt.Println(d.globalPat)
	}

	return nil
}

// 解析PMT表数据
func (d *TsDemuxer) readTsPMT(payload []byte, pHeader *TsHeader) error {

	// 获取临时PMT信息
	table_id := payload[0]
	section_syntax_indicator := payload[1] >> 7 & 0x1
	zero := payload[1] >> 6 & 0x1
	reserved1 := payload[1] >> 4 & 0x3
	section_length := uint16(payload[1]&0x0f)<<8 | uint16(payload[2])
	program_number := uint16(payload[3])<<8 | uint16(payload[4])
	reserved2 := payload[5] >> 6 & 0x3
	version_number := payload[5] >> 1 & 0x1f
	current_next_indicator := payload[5] & 0x1
	section_number := payload[6]
	last_section_number := payload[7]
	reserved3 := payload[8] >> 5 & 0x7
	PCR_PID := uint16(payload[8]&0x1f)<<8 | uint16(payload[9])
	reserved4 := payload[10] >> 4 & 0xf
	program_info_length := uint16(payload[10]&0xf)<<8 | uint16(payload[11])
	var CRC uint32
	var stream_count uint8

	// current_next_indicator 当前包无效
	// PMT是否有效,固定为1，表示这个PAT表可以用，如果为0则要等待下一个PMT表
	if current_next_indicator != 0x1 {
		return nil
	}

	// 检测固定位
	if zero != 0x0 {
		err := errors.NewError(ERROR_CODE_TS_PMT_READ_FAILED, "pmt parse error!zero!")
		fmt.Printf("%s \n", err.ErrMsg)
		return err
	}
	if section_syntax_indicator != 0x1 {
		err := errors.NewError(ERROR_CODE_TS_PMT_READ_FAILED, "不支持 section_syntax_indicator 为 0!")
		fmt.Printf("%s \n", err.ErrMsg)
		return err
	}

	// programInfo提取,未实现
	if program_info_length != 0x0 {

		// TODO
		// ...
	}

	// 有效负载总长度
	var plen int = 3 + int(section_length)
	CRC = uint32(payload[plen-4])<<24 | uint32(payload[plen-3])<<16 | uint32(payload[plen-2])<<8 | (uint32(payload[plen-1]) & 0xFF)

	// 提取循环部分字节数组
	var loopStartPos int = 12 + int(program_info_length)
	var loopLength int = int(section_length) - 13 - int(program_info_length)

	// 循环数据数组起始
	var pLoopData []byte = payload[loopStartPos : loopStartPos+loopLength]

	// 追加TS 分段语法缓存
	d.storeTsSectionData(section_number, pLoopData, pHeader)

	// 当前为最后分段，解析循环数据
	if section_number == last_section_number {

		var loopDataBuffer []byte = d.bufferMap[pHeader.PID]

		// 校验循环数据是否有变化，pat数据没变化，取消解析
		if d.globalPmt.pLoopData != nil && bytes.Equal(d.globalPmt.pLoopData, loopDataBuffer) {

			// 清空缓存数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
			return nil
		}

		var pos int = 0
		var streamcount uint8 = 0

		// 获取有多少流
		for pos < len(pLoopData) {
			es_info_length := (uint16(pLoopData[pos+3]&0xf) << 8) | uint16(pLoopData[pos+4])
			if es_info_length > 0 {
				pos += int(es_info_length)
			}
			pos += 5
			streamcount++
		}

		// 寄存流总数
		stream_count = streamcount

		var streams []TsStream = make([]TsStream, stream_count)

		pos = 0
		streamcount = 0
		for pos < len(pLoopData) {
			var s TsStream
			s.stream_type = pLoopData[pos]
			s.reserved1 = (pLoopData[pos+1] >> 5) & 0x7
			s.elementary_PID = uint16(pLoopData[pos+1]&0x1f)<<8 | uint16(pLoopData[pos+2])
			s.reserved2 = (pLoopData[pos+3] >> 4) & 0xf
			s.ES_info_length = uint16(pLoopData[pos+3]&0xf)<<8 | uint16(pLoopData[pos+4])

			if s.ES_info_length > 0 {

				// TODO 暂未解析
				pos += int(s.ES_info_length)
			}

			streams[streamcount] = s
			pos += 5
			streamcount++
		}

		d.globalPmt.table_id = table_id
		d.globalPmt.section_syntax_indicator = section_syntax_indicator
		d.globalPmt.zero = zero
		d.globalPmt.reserved1 = reserved1
		d.globalPmt.section_length = section_length
		d.globalPmt.program_number = program_number
		d.globalPmt.reserved2 = reserved2
		d.globalPmt.version_number = version_number
		d.globalPmt.current_next_indicator = current_next_indicator
		d.globalPmt.section_number = section_number
		d.globalPmt.last_section_number = last_section_number
		d.globalPmt.reserved3 = reserved3
		d.globalPmt.PCR_PID = PCR_PID
		d.globalPmt.reserved4 = reserved4
		d.globalPmt.program_info_length = program_info_length
		d.globalPmt.CRC = CRC
		d.globalPmt.stream_count = stream_count
		d.globalPmt.pLoopData = loopDataBuffer
		d.globalPmt.streams = streams

		var video_found bool = false
		var audio_found bool = false
		var i int

		// 设置视频、音频
		for i = 0; i < int(d.globalPmt.stream_count); i++ {

			// h.264编码对应0x1b
			// aac编码对应0x0f
			if d.globalPmt.streams[i].stream_type == 0x1b && !video_found {
				d.curVideoPID = int(d.globalPmt.streams[i].elementary_PID)
				video_found = true
			}
			if d.globalPmt.streams[i].stream_type == 0x0f && !audio_found {
				d.curAudioPID = int(d.globalPmt.streams[i].elementary_PID)
				audio_found = true
			}
		}

		fmt.Printf("识别到PMT表，PID：%d, streamcount:%d \n", pHeader.PID, stream_count)
		fmt.Printf("识别到当前视频流，PID：%d \n", d.curVideoPID)
		fmt.Printf("识别到当前音频流，PID：%d \n", d.curAudioPID)
		fmt.Println(d.globalPmt)
	}

	return nil
}

// 读取pes有效载荷，得到帧数据
func (d *TsDemuxer) readPesPayload(payload []byte, pHeader *TsHeader) error {

	// 缓存中已经存在帧数据
	if d.bufferMap[pHeader.PID] != nil {

		// ts包中含有新pes包头时
		if pHeader.payload_unit_start_indicator == 0x1 {

			// 高清节目帧长度超过PES_packet_length 16位显示极限
			// 此时PES_packet_length=0，每个帧以下一个新帧开始作为结束标志
			// ts包中含有新pes包头时，集中处理上一帧的缓存数据，并清空缓存
			if len(d.bufferMap[pHeader.PID]) > 0 {
				fmt.Printf("newstart curPesLen %d, PID: %d, buffersize: %d \n", d.curPesLen, pHeader.PID, len(d.bufferMap[pHeader.PID]))

				// 解析PES数据
				d.readPes(d.bufferMap[pHeader.PID], pHeader)

				// 清空旧数据
				d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
				d.curPesLen = -1

				fmt.Printf("clean:%d \n", len(d.bufferMap[pHeader.PID]))
			}
		}

	} else {

		if pHeader.payload_unit_start_indicator == 0x1 {

			// 创建缓存
			d.bufferMap[pHeader.PID] = make([]byte, 0)
		}
	}

	// 缓存中追加新数据
	//fmt.Printf("zhuijia : %d \n", len(payload))
	d.bufferMap[pHeader.PID] = append(d.bufferMap[pHeader.PID], payload...)

	// 没有总长度时，尝试根据已有信息读取总pes长度
	if d.curPesLen < 0 && len(d.bufferMap[pHeader.PID]) >= 6 {

		pesBuffer := d.bufferMap[pHeader.PID]

		// PES 包长度
		PES_packet_length := uint16(pesBuffer[4])<<8 | uint16(pesBuffer[5])

		// 加了pes头六个字节，得到pes包全长,否则为超长pes包，curPesLen为零
		if PES_packet_length != 0 {
			d.curPesLen = int(PES_packet_length) + 6
		} else {
			d.curPesLen = 0
		}
	}

	// 判断是否已经满足本帧的长度
	if d.curPesLen > 0 && d.curPesLen == len(d.bufferMap[pHeader.PID]) {

		fmt.Printf("autostop curPesLen %d, PID: %d, buffersize: %d \n", d.curPesLen, pHeader.PID, len(d.bufferMap[pHeader.PID]))

		// 解析PES数据
		d.readPes(d.bufferMap[pHeader.PID], pHeader)

		// 清空旧数据
		d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
		d.curPesLen = -1

		fmt.Printf("clean:%d \n", len(d.bufferMap[pHeader.PID]))
	}

	return nil
}

// PES包解析
func (d *TsDemuxer) readPes(pesBuffer []byte, pHeader *TsHeader) error {

	var tp TsPes

	tp.pes_start_code_prefix = uint32(pesBuffer[0])<<16 | uint32(pesBuffer[1])<<8 | uint32(pesBuffer[2])
	if tp.pes_start_code_prefix != 0x001 {
		err := errors.NewError(ERROR_CODE_TS_PES_READ_FAILED, "pes_start_code_prefix error!")
		fmt.Printf("%s \n", err.ErrMsg)

		fmt.Printf("%d (%d) >  stream_id: %d, pesStartPrefix: %d  PES_packet_length:%d  realLen:%d \n",
			pHeader.PID, pHeader.payload_unit_start_indicator, pesBuffer[3], tp.pes_start_code_prefix, tp.PES_packet_length, len(pesBuffer))

		return err
	}
	tp.stream_id = pesBuffer[3]
	tp.PES_packet_length = uint16(pesBuffer[4])<<8 | uint16(pesBuffer[5])
	tp.twobit_10 = pesBuffer[6] >> 6 & 0x3
	tp.PES_scrambling_control = pesBuffer[6] >> 4 & 0x3
	tp.PES_priority = pesBuffer[6] >> 3 & 0x1
	tp.data_alignment_indicator = pesBuffer[6] >> 2 & 0x1
	tp.copyright = pesBuffer[6] >> 1 & 0x1
	tp.original_or_copy = pesBuffer[6] & 0x1
	tp.PTS_DTS_flags = pesBuffer[7] >> 6 & 0x3
	tp.ESCR_flag = pesBuffer[7] >> 5 & 0x1
	tp.ES_rate_flag = pesBuffer[7] >> 4 & 0x1
	tp.DSM_trick_mode_flag = pesBuffer[7] >> 3 & 0x1
	tp.additional_copy_info_flag = pesBuffer[7] >> 2 & 0x1
	tp.PES_CRC_flag = pesBuffer[7] >> 1 & 0x1
	tp.PES_extension_flag = pesBuffer[7] & 0x1
	tp.PES_header_data_length = pesBuffer[8]

	// 可选域字节索引
	var opt_field_idx int = 9

	// PTS(presentation time stamp 显示时间标签)
	// DTS(decoding time stamp 解码时间标签)标志位
	if tp.PTS_DTS_flags == 0x2 {
		tp.PTS = (uint64(pesBuffer[opt_field_idx])>>1&0x7)<<30 |
			uint64(pesBuffer[opt_field_idx+1])<<22 |
			(uint64(pesBuffer[opt_field_idx+2])>>1&0x7f)<<15 |
			uint64(pesBuffer[opt_field_idx+3])<<7 |
			(uint64(pesBuffer[opt_field_idx+4]) >> 1 & 0x7f)
		opt_field_idx += 5
		tp.DTS = tp.PTS
	} else if tp.PTS_DTS_flags == 0x3 {
		tp.PTS = (uint64(pesBuffer[opt_field_idx])>>1&0x7)<<30 |
			uint64(pesBuffer[opt_field_idx+1])<<22 |
			(uint64(pesBuffer[opt_field_idx+2])>>1&0x7f)<<15 |
			uint64(pesBuffer[opt_field_idx+3])<<7 |
			(uint64(pesBuffer[opt_field_idx+4]) >> 1 & 0x7f)

		tp.DTS = (uint64(pesBuffer[opt_field_idx+5])>>1&0x7)<<30 |
			uint64(pesBuffer[opt_field_idx+6])<<22 |
			(uint64(pesBuffer[opt_field_idx+7])>>1&0x7f)<<15 |
			uint64(pesBuffer[opt_field_idx+8])<<7 |
			(uint64(pesBuffer[opt_field_idx+9]) >> 1 & 0x7f)
		opt_field_idx += 10
	}

	//	fmt.Printf("PTS:%d, DTS:%d \n ",tp.PTS,tp.DTS )

	fmt.Printf("%d (%d) >  stream_id: %d, pesStartPrefix: %d  PES_packet_length:%d  realLen:%d \n",
		pHeader.PID, pHeader.payload_unit_start_indicator, pesBuffer[3], tp.pes_start_code_prefix, tp.PES_packet_length, len(pesBuffer))

	return nil
}

// 追加TS 分段语法缓存
func (d *TsDemuxer) storeTsSectionData(section_number uint8, pLoopData []byte, pHeader *TsHeader) {

	// 当前段是第一个分段
	if section_number == 0x00 {

		if d.bufferMap[pHeader.PID] == nil {

			// ts包字节 切片
			d.bufferMap[pHeader.PID] = make([]byte, 0)

		} else {

			// 清空旧数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
		}

		// 向缓存追加数据
		d.bufferMap[pHeader.PID] = append(d.bufferMap[pHeader.PID], pLoopData...)

	} else {

		if d.bufferMap[pHeader.PID] != nil {

			// 向缓存追加数据
			d.bufferMap[pHeader.PID] = append(d.bufferMap[pHeader.PID], pLoopData...)
		}
	}
}
