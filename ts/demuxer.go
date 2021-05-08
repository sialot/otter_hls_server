package ts

import (
	"bytes"
	"fmt"

	errors "../errors"
)

// TsPkgSize TS包字节数
const TsPkgSize int = 188

// TsReloadNum 预加载包数量
const TsReloadNum int = 100000

// header Ts头
type header struct {
	syncByte                   uint8  //8 同步字节：固定为0x47;
	transportErrorIndicator    uint8  //1 传输错误标志：‘1’表示在相关的传输包中至少有一个不可纠正的错误位。
	payloadUnitStartIndicator  uint8  //1 负载起始标志：在前4个字节之后会有一个调整字节，其的数值为后面调整字段的长度length。
	transportPriority          uint8  //1 传输优先级标志
	PID                        uint16 //13 PID
	transportScramblingControl uint8  //2 加扰控制标志：表示TS流分组有效负载的加密模式。空包为‘00’
	adaptationFieldControl     uint8  //2 适配域控制标志‘00’为ISO/IEC未来使用保留；
	continuityCounter          uint8  //4 连续性计数器
	adaptaionFieldLength       uint8  //8 适配域长度
}

// patProgram pat 中的 Program
type patProgram struct {
	programNumber uint16 //16 节目号
	reserved      uint8  //3 保留字段，固定为111
	PID           uint16 //16 节目号对应内容的PID值
}

// pat 表
type pat struct {
	tableID                uint8  //8 pat表固定为0x00
	sectionSyntaxIndicator uint8  //1 段语法标志位，固定为1
	zero                   uint8  //1 固定为0
	reserved1              uint8  //2 保留字段，固定为11
	sectionLength          uint16 //12 表示这个字节后面数据的长度,包括 CRC信息
	transportstreamID      uint16 //16 该传输流的ID
	reserved2              uint8  //2 保留字段，固定为11
	versionNumber          uint8  //5 版本号，固定为00000,有变化则版本号加1
	currentNextIndicator   uint8  //1 pat是否有效,固定为1，表示这个pat表可以用，如果为0则要等待下一个pat表
	sectionNumber          uint8  //8 分段号码,最多256个分段
	lastSectionNumber      uint8  //8 最后一个分段的号码
	networkPID             uint16 //16 网络PID
	CRC                    uint32 //32 CRC校验码
	programCount           uint8  //8 节目数量
	pLoopData              []byte //循环数据
	programs               []patProgram
}

// stream 视频流信息结构体
type stream struct {
	streamType    uint8  //8 流类型  h.264编码对应0x1b;aac编码对应0x0f;mp3编码对应0x03
	reserved1     uint8  //3 保留字段，固定为111
	elementaryPID uint16 //13 元素PID,与streamType对应的PID
	reserved2     uint8  //4 保留字段，固定为1111
	ESInfoLength  uint16 //12  描述信息，指定为0x000表示没有
}

// pmt 表
type pmt struct {
	tableID                uint8    //8 pat表固定为0x00
	sectionSyntaxIndicator uint8    //1 段语法标志位，固定为1
	zero                   uint8    //1 固定为0
	reserved1              uint8    //2  保留字段，固定为11
	sectionLength          uint16   //12 表示这个字节后面数据的长度,包括 CRC信息
	programNumber          uint16   //16 频道号码，表示当前的pmt关联到的频道，取值0x0001
	reserved2              uint8    //2 保留字段，固定为11
	versionNumber          uint8    //5 版本号，固定为00000，如果pat有变化则版本号加1
	currentNextIndicator   uint8    //1 是否有效
	sectionNumber          uint8    //8 分段号码
	lastSectionNumber      uint8    //8 最后一个分段的号码
	reserved3              uint8    //3 保留字段，固定为111
	PcrPID                 uint16   //13 PCR_PID PCR(节目参考时钟)所在TS分组的PID，指定为视频PID
	reserved4              uint8    //4 保留字段固定为 1111
	programInfoLength      uint16   //12 节目描述信息，指定为0x000表示没有
	CRC                    uint32   //32 CRC校验码
	streamCount            uint8    //8 流总数
	pLoopData              []byte   //循环数据
	streams                []stream // 流数据
}

// Pes pes数据结构体
type Pes struct {
	pesStartCodePrefix     uint32 //24 起始码，固定必须是'0000 0000 0000 0000 0000 0001' (0x000001)。用于标识包的开始。
	streamID               uint8  //8 流ID
	PESPacketLength        uint16 //16 PES包的长度
	twobit10               uint8  //2 固定两位分割bit 0x2
	PESScramblingControl   uint8  //2 字段指示 PES 包有效载荷的加扰方式; PES 包头，其中包括任选字段只要存在，应不加扰。00 不加扰
	PESPriority            uint8  //1 指示在此 PES 包中该有效载荷的优先级。
	dataAlignmentIndicator uint8  //1 数据校准标志
	copyright              uint8  //1 版权保护标志
	originalOrCopy         uint8  //1 是否为复制
	PtsDtsFlags            uint8  //2 PTS(presentation time stamp 显示时间标签),DTS(decoding time stamp 解码时间标签)标志位
	ESCRFlag               uint8  //1 置于‘1’时指示 PES 包头中 ESCR 基准字段和 ESCR 扩展字段均存在。
	ESRateFlag             uint8  //1 置于‘1’时指示 PES 包头中 ESRate 字段存在。
	DSMTrickModeFlag       uint8  //1 特技方式
	additionalCopyInfoFlag uint8  //1 置于‘1’时指示 additionalCopyInfo 存在。
	PESCRCFlag             uint8  //1 置于‘1’时指示 PES 包中 CRC 字段存在。
	PESExtensionFlag       uint8  //1 置于‘1’时指示 PES 包头中扩展字段存在。置于‘0’时指示此字段不存在
	PESHeaderDataLength    uint8  //8 指示在此PES包头中包含的由任选字段和任意填充字节所占据的字节总数。
	PTS                    int64  //33 PTS(presentation time stamp 显示时间标签)
	DTS                    int64  //33 DTS(decoding time stamp 解码时间标签)标志位
	ESCRBase               uint64 //33 基本流时钟参考
	ESCRExtension          uint16 //9 基本流时钟参考
	ESRate                 uint32 //22 ES 速率（基本流速率）
	trickModeControl       uint8  //3 3 比特字段，指示适用于相关视频流的特技方式
	fieldID                uint8  //2 2 比特字段，指示哪些字段应予显示
	intraSliceRefresh      uint8  //1 1 比特标志，置于‘1’时指示此 PES 包中视频数据的编码截面间可能存在丢失宏块
	frequencyTruncation    uint8  //2 指示在此 PES 包中编码视频数据时曾经使用的受限系数集
	repCntrl               uint8  //5 指示交错图像中每个字段应予显示的次数，或者连续图像应予显示的次数
	additionalCopyInfo     uint8  //7 此 7 比特字段包含与版权信息有关的专用数据
	previousPESPacketCRC   uint16 //16 包含产生解码器中 16 寄存器零输出的 CRC 值
	PkgOffset              uint64 // pes开始位置所处的文件偏移量
	ptime                  int64
	dtime                  int64
	PID                    uint16
}

// Demuxer TS解封装器
type Demuxer struct {
	globalpat     pat               // 全局pat表
	globalpmt     pmt               // 全局pmt表
	bufferMap     map[uint16][]byte // 全局ts buffer临时存储，key PID,值 byte数据切片
	curPesLen     int               // 当前pes结束长度
	curVideoPID   int
	curAudioPID   int
	curOffset     uint64
}

// Init 初始化解封装器
func (d *Demuxer) Init() {
	d.bufferMap = make(map[uint16][]byte)
	d.curPesLen = -1
	d.curVideoPID = -1
	d.curAudioPID = -1
	d.curOffset = 0
}

// DemuxPkg 解封装
func (d *Demuxer) DemuxPkg(pKgBuf []byte) (*Pes, error) {

	// 新包头记录当前包头的偏移量
	d.curOffset += uint64(TsPkgSize)

	// check包长度
	if 188 != len(pKgBuf) {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsPackage length is not 188!")
		return nil, err
	}

	// 获取包头
	header, err := d.readTsHeader(pKgBuf)
	if err != nil {
		return nil, err
	}

	// 获取适配域
	adpReadErr := d.readAdaptionField(pKgBuf, header)
	if adpReadErr != nil {
		return nil, adpReadErr
	}

	// 获取有效载荷, adaptationFieldControl 01,11 代表有有效载荷
	if header.adaptationFieldControl == 0x01 || header.adaptationFieldControl == 0x03 {
		pesResult, payloadReadErr := d.readPayload(pKgBuf, header)
		if payloadReadErr != nil {
			return nil, payloadReadErr
		}
		if pesResult != nil {
			return pesResult, nil
		}
	}

	return nil, nil
}

// 解析TS包头
func (d *Demuxer) readTsHeader(pKgBuf []byte) (*header, error) {
	var header header
	header.syncByte = pKgBuf[0]

	// 不是有效的ts包，抛弃
	if header.syncByte != 0x47 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "TsHeader read failed!")
		return nil, err
	}

	header.transportErrorIndicator = pKgBuf[1] >> 7
	header.payloadUnitStartIndicator = pKgBuf[1] >> 6 & 0x01
	header.transportPriority = pKgBuf[1] >> 5 & 0x01
	header.PID = uint16(pKgBuf[1]&0x1f)<<8 | uint16(pKgBuf[2])
	header.transportScramblingControl = pKgBuf[3] >> 6
	header.adaptationFieldControl = pKgBuf[3] >> 4 & 0x03
	header.continuityCounter = pKgBuf[3] & 0x0f

	return &header, nil
}

// 解析适配域
func (d *Demuxer) readAdaptionField(pKgBuf []byte, pHeader *header) error {

	if pHeader.adaptationFieldControl == 0x2 || pHeader.adaptationFieldControl == 0x3 {
		pHeader.adaptaionFieldLength = pKgBuf[4]
	}

	//  TODO 具体解析未实现
	return nil
}

// 解析有效载荷
func (d *Demuxer) readPayload(pKgBuf []byte, pHeader *header) (*Pes, error) {

	// 负载信息起始索引
	var start int = 4

	// 同时存在负载和适配域
	if pHeader.adaptationFieldControl == 0x3 {
		start = start + 1 + int(pHeader.adaptaionFieldLength)
	}

	// 看是否为pat信息
	if pHeader.PID == 0x0 {

		// 对于PSI,payloadUnitStartIndicator 为1时
		// 有效载荷开始的位置应再偏移 1 个字节,pointer_field。
		if pHeader.payloadUnitStartIndicator == 0x01 {
			start = start + 1
		}

		err := d.readpat(pKgBuf[start:len(pKgBuf)], pHeader)
		if err != nil {
			return nil, err
		}
	}

	// 是否为 BAT/SDT 信息
	if pHeader.PID == 0x11 {
		//  TODO 具体解析未实现
		//  ...
		//  ...
	}

	// 仍然未找到有效的视频流
	if(d.curVideoPID == -1){

		// 同时解析每一个program，存在有效的pmt表后，第一个program会更新curVideoPID，后续program将被忽略
		var i int
		for i = 0; i < len(d.globalpat.programs); i ++ {

			// 看是否为pmt信息
			if d.globalpat.programs[i].PID == pHeader.PID {

				// 对于PSI,payloadUnitStartIndicator 为1时
				// 有效载荷开始的位置应再偏移 1 个字节,pointer_field。
				if pHeader.payloadUnitStartIndicator == 0x01 {
					start = start + 1
				}
				err := d.readpmt(pKgBuf[start:len(pKgBuf)], pHeader)

				if err != nil {
					return nil, err
				}
			}

		}

	} else {

		// 切片只需要视频信息
		if int(pHeader.PID) == d.curVideoPID {

			// 解析PES数据
			pesResult, err := d.readPesPayload(pKgBuf[start:len(pKgBuf)], pHeader)

			if err != nil {
				return nil, err
			}

			if pesResult != nil {
				return pesResult, nil
			}
		}
	}

	return nil, nil
}

// 解析pat表数据
func (d *Demuxer) readpat(payload []byte, pHeader *header) error {

	// 获取临时pat表
	tableID := payload[0]
	sectionSyntaxIndicator := payload[1] >> 7 & 0x1
	zero := payload[1] >> 6 & 0x1
	reserved1 := payload[1] >> 4 & 0x3
	sectionLength := uint16(payload[1]&0x0f)<<8 | uint16(payload[2])
	transportstreamID := uint16(payload[3])<<8 | uint16(payload[4])
	reserved2 := payload[5] >> 6 & 0x3
	versionNumber := payload[5] >> 1 & 0x1f
	currentNextIndicator := payload[5] & 0x1
	sectionNumber := payload[6]
	lastSectionNumber := payload[7]
	var networkPID uint16
	var CRC uint32
	var programCount uint8

	// currentNextIndicator 当前包无效
	// pat是否有效,固定为1，表示这个pat表可以用，如果为0则要等待下一个pat表
	if currentNextIndicator != 0x1 {
		return nil
	}

	// 检测三个固定位
	if tableID != 0x00 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "pat parse error!tableID!")
		return err
	}
	if zero != 0x0 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "pat parse error!zero!")
		return err
	}
	if sectionSyntaxIndicator != 0x1 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "不支持 sectionSyntaxIndicator 为 0!")
		return err
	}

	// 有效负载总长度
	var plen int = 3 + int(sectionLength)
	CRC = uint32(payload[plen-4])<<24 | uint32(payload[plen-3])<<16 | uint32(payload[plen-2])<<8 | (uint32(payload[plen-1]) & 0xFF)

	// 提取循环部分字节数组
	var loopStartPos int = 8
	var loopLength int = int(sectionLength) - 9

	// 提取循环数据，开始为loopStartPos， 结束为 loopStartPos + loopLength
	var pLoopData []byte = payload[loopStartPos : loopStartPos+loopLength]

	// 追加TS 分段语法缓存
	d.storeTsSectionData(sectionNumber, pLoopData, pHeader)

	// 当前为最后分段，解析循环数据
	if sectionNumber == lastSectionNumber {

		var loopDataBuffer []byte = d.bufferMap[pHeader.PID]

		// 校验循环数据是否有变化，pat数据没变化，取消解析
		if d.globalpat.pLoopData != nil && bytes.Equal(d.globalpat.pLoopData, loopDataBuffer) {

			// 清空缓存数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
			return nil
		}

		// 校验循环数据长度，如果不为4的整数倍，数据有错误
		if len(loopDataBuffer)%4 != 0 {

			err := errors.NewError(errors.ErrorCodeDemuxFailed, "pat parse error!pat.loopData.length!")
			return err
		}

		// 根据缓存计算program数
		programCount = uint8(len(loopDataBuffer) / 4)

		var programs []patProgram = make([]patProgram, programCount)

		var i int
		for i = 0; i < len(loopDataBuffer); i += 4 {

			var programNumber uint16 = uint16(loopDataBuffer[i]&0xff)<<8 | uint16(loopDataBuffer[i+1]&0xff)

			// 0x00 是NIT
			if programNumber == 0x00 {
				networkPID = uint16(loopDataBuffer[i+2]&0x1f)<<8 | uint16(loopDataBuffer[i+3]&0xff)
			} else {
				var prg patProgram
				prg.programNumber = programNumber
				prg.reserved = loopDataBuffer[i+2] >> 5 & 0x3
				prg.PID = uint16(loopDataBuffer[i+2]&0x1f)<<8 | uint16(loopDataBuffer[i+3]&0xff)
				programs[i/4] = prg
			}
		}

		// 提交临时pat表到全局pat表
		d.globalpat.tableID = tableID
		d.globalpat.sectionSyntaxIndicator = sectionSyntaxIndicator
		d.globalpat.zero = zero
		d.globalpat.reserved1 = reserved1
		d.globalpat.sectionLength = sectionLength
		d.globalpat.transportstreamID = transportstreamID
		d.globalpat.reserved2 = reserved2
		d.globalpat.versionNumber = versionNumber
		d.globalpat.currentNextIndicator = currentNextIndicator
		d.globalpat.sectionNumber = sectionNumber
		d.globalpat.lastSectionNumber = lastSectionNumber
		d.globalpat.networkPID = networkPID
		d.globalpat.CRC = CRC
		d.globalpat.programCount = programCount
		d.globalpat.pLoopData = loopDataBuffer
		d.globalpat.programs = programs

		Log.Debug("识别到pat表，PID：" + fmt.Sprint(pHeader.PID))
		Log.Debug(fmt.Sprint(d.globalpat))
	}

	return nil
}

// 解析pmt表数据
func (d *Demuxer) readpmt(payload []byte, pHeader *header) error {

	// 获取临时pmt信息
	tableID := payload[0]
	sectionSyntaxIndicator := payload[1] >> 7 & 0x1
	zero := payload[1] >> 6 & 0x1
	reserved1 := payload[1] >> 4 & 0x3
	sectionLength := uint16(payload[1]&0x0f)<<8 | uint16(payload[2])
	programNumber := uint16(payload[3])<<8 | uint16(payload[4])
	reserved2 := payload[5] >> 6 & 0x3
	versionNumber := payload[5] >> 1 & 0x1f
	currentNextIndicator := payload[5] & 0x1
	sectionNumber := payload[6]
	lastSectionNumber := payload[7]
	reserved3 := payload[8] >> 5 & 0x7
	PcrPID := uint16(payload[8]&0x1f)<<8 | uint16(payload[9])
	reserved4 := payload[10] >> 4 & 0xf
	programInfoLength := uint16(payload[10]&0xf)<<8 | uint16(payload[11])
	var CRC uint32
	var streamCount uint8

	// currentNextIndicator 当前包无效
	// pmt是否有效,固定为1，表示这个pat表可以用，如果为0则要等待下一个pmt表
	if currentNextIndicator != 0x1 {
		return nil
	}

	// 检测固定位
	if zero != 0x0 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "pmt parse error!zero!")
		return err
	}
	if sectionSyntaxIndicator != 0x1 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "不支持 sectionSyntaxIndicator 为 0!")
		return err
	}

	// programInfo提取,未实现
	if programInfoLength != 0x0 {

		// TODO
		// ...
	}

	// 有效负载总长度
	var plen int = 3 + int(sectionLength)
	CRC = uint32(payload[plen-4])<<24 | uint32(payload[plen-3])<<16 | uint32(payload[plen-2])<<8 | (uint32(payload[plen-1]) & 0xFF)

	// 提取循环部分字节数组
	var loopStartPos int = 12 + int(programInfoLength)
	var loopLength int = int(sectionLength) - 13 - int(programInfoLength)

	// 循环数据数组起始
	var pLoopData []byte = payload[loopStartPos : loopStartPos+loopLength]

	// 追加TS 分段语法缓存
	d.storeTsSectionData(sectionNumber, pLoopData, pHeader)

	// 当前为最后分段，解析循环数据
	if sectionNumber == lastSectionNumber {

		var loopDataBuffer []byte = d.bufferMap[pHeader.PID]

		// 校验循环数据是否有变化，pat数据没变化，取消解析
		if d.globalpmt.pLoopData != nil && bytes.Equal(d.globalpmt.pLoopData, loopDataBuffer) {

			// 清空缓存数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
			return nil
		}

		var pos int = 0
		var streamcount uint8 = 0

		// 获取有多少流
		for pos < len(pLoopData) {
			ESInfoLength := (uint16(pLoopData[pos+3]&0xf) << 8) | uint16(pLoopData[pos+4])
			if ESInfoLength > 0 {
				pos += int(ESInfoLength)
			}
			pos += 5
			streamcount++
		}

		// 寄存流总数
		streamCount = streamcount

		var streams []stream = make([]stream, streamCount)

		pos = 0
		streamcount = 0
		for pos < len(pLoopData) {
			var s stream
			s.streamType = pLoopData[pos]
			s.reserved1 = (pLoopData[pos+1] >> 5) & 0x7
			s.elementaryPID = uint16(pLoopData[pos+1]&0x1f)<<8 | uint16(pLoopData[pos+2])
			s.reserved2 = (pLoopData[pos+3] >> 4) & 0xf
			s.ESInfoLength = uint16(pLoopData[pos+3]&0xf)<<8 | uint16(pLoopData[pos+4])

			if s.ESInfoLength > 0 {

				// TODO 暂未解析
				pos += int(s.ESInfoLength)
			}

			streams[streamcount] = s
			pos += 5
			streamcount++
		}

		d.globalpmt.tableID = tableID
		d.globalpmt.sectionSyntaxIndicator = sectionSyntaxIndicator
		d.globalpmt.zero = zero
		d.globalpmt.reserved1 = reserved1
		d.globalpmt.sectionLength = sectionLength
		d.globalpmt.programNumber = programNumber
		d.globalpmt.reserved2 = reserved2
		d.globalpmt.versionNumber = versionNumber
		d.globalpmt.currentNextIndicator = currentNextIndicator
		d.globalpmt.sectionNumber = sectionNumber
		d.globalpmt.lastSectionNumber = lastSectionNumber
		d.globalpmt.reserved3 = reserved3
		d.globalpmt.PcrPID = PcrPID
		d.globalpmt.reserved4 = reserved4
		d.globalpmt.programInfoLength = programInfoLength
		d.globalpmt.CRC = CRC
		d.globalpmt.streamCount = streamCount
		d.globalpmt.pLoopData = loopDataBuffer
		d.globalpmt.streams = streams

		var isVideoFound bool = false
		var isAudioFound bool = false
		var i int

		// 设置视频、音频
		for i = 0; i < int(d.globalpmt.streamCount); i++ {

			// h.264编码对应0x1b
			// aac编码对应0x0f
			if d.globalpmt.streams[i].streamType == 0x1b && !isVideoFound {
				d.curVideoPID = int(d.globalpmt.streams[i].elementaryPID)
				isVideoFound = true
			}
			if d.globalpmt.streams[i].streamType == 0x0f && !isAudioFound {
				d.curAudioPID = int(d.globalpmt.streams[i].elementaryPID)
				isAudioFound = true
			}
		}

		Log.Debug("识别到pmt表，PID：" + fmt.Sprint(pHeader.PID) + ", streamcount: " + fmt.Sprint(streamCount))
		Log.Debug("识别到当前视频流，PID：" + fmt.Sprint(d.curVideoPID))
		Log.Debug("识别到当前音频流，PID：" + fmt.Sprint(d.curAudioPID))
		Log.Debug(fmt.Sprint(d.globalpmt))
	}

	return nil
}

// 读取pes有效载荷，得到帧数据
func (d *Demuxer) readPesPayload(payload []byte, pHeader *header) (*Pes, error) {

	var pesResult *Pes
	var err error

	// ts包中含有新pes包头时
	if pHeader.payloadUnitStartIndicator == 0x1 {

		if len(d.bufferMap[pHeader.PID]) > 0 {

			// 解析PES数据
			pesResult, err = d.readPes(d.bufferMap[pHeader.PID], pHeader)

			// 清空旧数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
			d.curPesLen = -1

			if err != nil {
				return nil, err
			}
		}

		d.bufferMap[pHeader.PID] = append(d.bufferMap[pHeader.PID], payload...)

	} else {

		d.bufferMap[pHeader.PID] = append(d.bufferMap[pHeader.PID], payload...)

		// 判断是否已经满足本帧的长度
		if d.curPesLen > 0 && d.curPesLen == len(d.bufferMap[pHeader.PID]) {

			// 解析PES数据
			pesResult, err = d.readPes(d.bufferMap[pHeader.PID], pHeader)

			// 清空旧数据
			d.bufferMap[pHeader.PID] = d.bufferMap[pHeader.PID][0:0]
			d.curPesLen = -1

			if err != nil {
				return nil, err
			}
		}
	}

	if pesResult != nil {
		return pesResult, nil
	}
	return nil, nil
}

// PES包解析
func (d *Demuxer) readPes(pesBuffer []byte, pHeader *header) (*Pes, error) {

	var tp Pes

	tp.pesStartCodePrefix = uint32(pesBuffer[0])<<16 | uint32(pesBuffer[1])<<8 | uint32(pesBuffer[2])

	if tp.pesStartCodePrefix != 0x001 {
		err := errors.NewError(errors.ErrorCodeDemuxFailed, "pesStartCodePrefix error!")
		return nil, err
	}
	tp.streamID = pesBuffer[3]
	tp.PESPacketLength = uint16(pesBuffer[4])<<8 | uint16(pesBuffer[5])
	tp.twobit10 = pesBuffer[6] >> 6 & 0x3
	tp.PESScramblingControl = pesBuffer[6] >> 4 & 0x3
	tp.PESPriority = pesBuffer[6] >> 3 & 0x1
	tp.dataAlignmentIndicator = pesBuffer[6] >> 2 & 0x1
	tp.copyright = pesBuffer[6] >> 1 & 0x1
	tp.originalOrCopy = pesBuffer[6] & 0x1
	tp.PtsDtsFlags = pesBuffer[7] >> 6 & 0x3
	tp.ESCRFlag = pesBuffer[7] >> 5 & 0x1
	tp.ESRateFlag = pesBuffer[7] >> 4 & 0x1
	tp.DSMTrickModeFlag = pesBuffer[7] >> 3 & 0x1
	tp.additionalCopyInfoFlag = pesBuffer[7] >> 2 & 0x1
	tp.PESCRCFlag = pesBuffer[7] >> 1 & 0x1
	tp.PESExtensionFlag = pesBuffer[7] & 0x1
	tp.PESHeaderDataLength = pesBuffer[8]

	// 可选域字节索引
	var optFieldIDx int = 9

	// PTS(presentation time stamp 显示时间标签)
	// DTS(decoding time stamp 解码时间标签)标志位
	if tp.PtsDtsFlags == 0x2 {
		tp.PTS = (int64(pesBuffer[optFieldIDx])>>1&0x7)<<30 |
			int64(pesBuffer[optFieldIDx+1])<<22 |
			(int64(pesBuffer[optFieldIDx+2])>>1&0x7f)<<15 |
			int64(pesBuffer[optFieldIDx+3])<<7 |
			(int64(pesBuffer[optFieldIDx+4]) >> 1 & 0x7f)
		optFieldIDx += 5
		tp.DTS = tp.PTS
	} else if tp.PtsDtsFlags == 0x3 {
		tp.PTS = (int64(pesBuffer[optFieldIDx])>>1&0x7)<<30 |
			int64(pesBuffer[optFieldIDx+1])<<22 |
			(int64(pesBuffer[optFieldIDx+2])>>1&0x7f)<<15 |
			int64(pesBuffer[optFieldIDx+3])<<7 |
			(int64(pesBuffer[optFieldIDx+4]) >> 1 & 0x7f)

		tp.DTS = (int64(pesBuffer[optFieldIDx+5])>>1&0x7)<<30 |
			int64(pesBuffer[optFieldIDx+6])<<22 |
			(int64(pesBuffer[optFieldIDx+7])>>1&0x7f)<<15 |
			int64(pesBuffer[optFieldIDx+8])<<7 |
			(int64(pesBuffer[optFieldIDx+9]) >> 1 & 0x7f)
		optFieldIDx += 10
	}

	tp.PkgOffset = d.curOffset
	tp.ptime = tp.PTS / 90
	tp.dtime = tp.DTS / 90
	tp.PID = pHeader.PID
	return &tp, nil
}

// 追加TS 分段语法缓存
func (d *Demuxer) storeTsSectionData(sectionNumber uint8, pLoopData []byte, pHeader *header) {

	// 当前段是第一个分段
	if sectionNumber == 0x00 {

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
