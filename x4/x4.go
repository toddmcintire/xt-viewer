package x4

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

//56 bytes
type Header struct {
	mark string
	version uint16
	PageCount uint16
	readDirection uint8
	hasMetaData uint8
	hasThumbnails uint8
	hasChapters uint8
	currentPage uint32
	MetadataOffset uint64
	IndexOffset uint64
	dataOffset uint64
	thumbnailOffset uint64
	chapterOffset uint64
}

type ReadDirection uint8

const (
	LR ReadDirection = iota
	RL
	TB
)

//256 bytes (optional, at metadataOffset)
type Metadata struct {
	title string
	author string
	publisher string
	language string
	createTime uint32
	coverPage uint16
	ChapterCount uint16
	reserved uint64
}

// n * 96 bytes (optional, at chapterOffset)
type Chapter struct {
	chapterName string
	startPage uint16
	endPage uint16
	reserved1 uint32
	reserved2 uint32
	reserved3 uint32
}

// pageCount * 16 bytes (at indexOffset)
//thumbnail area (optional at thumbOffset after pageData)
type Page struct {
	offset uint64
	size uint32
	width uint16
	height uint16
}

type XTG struct {
	mark string
	width uint16
	height uint16
	colorMode uint8
	compression uint8
	dataSize uint32
	md5 uint64
	Data [48000]byte
}

type XTH struct {
	mark string
	width uint16
	height uint16
	colorMode uint8
	compression uint8
	dataSize uint32
	md5 uint64
	Data [96000]byte
}

func GetXTCHeader(filePT *os.File) (Header, error){
	var header Header
	headerBuffer := make([]byte, 56)

	bufferReadLen, err := filePT.ReadAt(headerBuffer, 0)
	if err != nil && bufferReadLen != 56 {
		return Header{}, fmt.Errorf("buffer read error: %v", err)
	}

	header.mark = string(headerBuffer[0:4])
	header.version = binary.LittleEndian.Uint16(headerBuffer[4:6])
	header.PageCount = binary.LittleEndian.Uint16(headerBuffer[6:8])
	header.readDirection = headerBuffer[8]
	header.hasMetaData = headerBuffer[9]
	header.hasThumbnails = headerBuffer[10]
	header.hasChapters = headerBuffer[11]
	header.currentPage = binary.LittleEndian.Uint32(headerBuffer[12:16])
	header.MetadataOffset = binary.LittleEndian.Uint64(headerBuffer[16:24])
	header.IndexOffset = binary.LittleEndian.Uint64(headerBuffer[24:32])
	header.dataOffset = binary.LittleEndian.Uint64(headerBuffer[32:40])
	header.thumbnailOffset = binary.LittleEndian.Uint64(headerBuffer[40:48])
	header.chapterOffset = binary.LittleEndian.Uint64(headerBuffer[48:56])
	
	if header.hasMetaData == 0 {
		return Header{}, errors.New("no metadata")
	}

	return header, nil
}

func GetXTCMetadata(filePT *os.File, offset uint64) (Metadata, error) {

	var metadata Metadata
	metadataBuffer := make([]byte, 256)

	bufferReadLen, err := filePT.ReadAt(metadataBuffer, int64(offset))
	if err != nil && bufferReadLen != 256 {
		return Metadata{}, fmt.Errorf("%v", err)
	}	

	metadata.title = string(metadataBuffer[:128])
	metadata.author = string(metadataBuffer[128:192])
	metadata.publisher = string(metadataBuffer[192:224])
	metadata.language = string(metadataBuffer[224:240])
	metadata.createTime = binary.LittleEndian.Uint32(metadataBuffer[240:244])
	metadata.coverPage = binary.LittleEndian.Uint16(metadataBuffer[244:246])
	metadata.ChapterCount = binary.LittleEndian.Uint16(metadataBuffer[246:248])
	metadata.reserved = binary.LittleEndian.Uint64(metadataBuffer[248:256])

	return metadata, nil
}

func getXTCChapter(filePT *os.File, offset uint64, count uint16) ([]Chapter, error) {
	var chapters []Chapter

	for i:=0; i<=int(count); i++ {
		chapterBuffer := make([]byte, 96)
		var chapter Chapter

		bufferReadLen, err := filePT.ReadAt(chapterBuffer, int64(offset))
		if err != nil && bufferReadLen != 96 {
			return []Chapter{}, fmt.Errorf("%v", err)
		}

		chapter.chapterName = string(chapterBuffer[:80])
		chapter.startPage = binary.LittleEndian.Uint16(chapterBuffer[80:82])
		chapter.endPage = binary.LittleEndian.Uint16(chapterBuffer[82:84])
		//skipping reserved

		chapters = append(chapters, chapter)
	}	

	return chapters, nil
}

func GetXTCPage(filePT *os.File, offset uint64, count uint16) ([]Page, error) {
	var pages []Page

	for i:=0; i<int(count); i++ {
		pageBuffer := make([]byte, 16)
		var page Page

		bufferReadLen, err := filePT.ReadAt(pageBuffer, int64(offset))
		if err != nil && bufferReadLen != 16 {
			return []Page{}, fmt.Errorf("%v", err)
		}

		page.offset = binary.LittleEndian.Uint64(pageBuffer[0:8])
		page.size = binary.LittleEndian.Uint32(pageBuffer[8:12])
		page.width = binary.LittleEndian.Uint16(pageBuffer[12:14])
		page.height = binary.LittleEndian.Uint16(pageBuffer[14:16])

		pages = append(pages, page)

		//increment offset
		offset += 16
	}
	return pages, nil
}

func GetXTCPages(pages []Page, filePT *os.File) (any, error) {
	//get page version from first page
	versionBuffer := make([]byte, 4)
	_, verBufErr := filePT.ReadAt(versionBuffer, int64(pages[0].offset))
	if verBufErr != nil {
		return nil, verBufErr
 	}

	//based on page version, return correct version
	if string(versionBuffer) == "XTG\x00" {
		var pictures []XTG

		for _, v := range pages {
			var pictureData XTG	
			pageDataBuffer := make([]byte, 48022)

			bufferReadLen, err := filePT.ReadAt(pageDataBuffer, int64(v.offset))
			if err != nil && bufferReadLen != 48022 {
				return []XTG{}, fmt.Errorf("%v", err)
			}

			pictureData.mark = string(pageDataBuffer[0:4])
			pictureData.width = binary.LittleEndian.Uint16(pageDataBuffer[4:6])
			pictureData.height = binary.LittleEndian.Uint16(pageDataBuffer[6:8])
			pictureData.colorMode = pageDataBuffer[8]
			pictureData.compression = pageDataBuffer[9]
			pictureData.dataSize = binary.LittleEndian.Uint32(pageDataBuffer[10:14])
			pictureData.md5 = binary.LittleEndian.Uint64(pageDataBuffer[14:22])
			pictureData.Data = [48000]byte(pageDataBuffer[22:])

			pictures = append(pictures, pictureData)
		}

		return pictures, nil

	} else if string(versionBuffer) == "XTH\x00" {
		var pictures []XTH

		for _, v := range pages {
			var pictureData XTH	
			pageDataBuffer := make([]byte, 96022)

			bufferReadLen, err := filePT.ReadAt(pageDataBuffer, int64(v.offset))
			if err != nil && bufferReadLen != 48022 {
				return []XTH{}, fmt.Errorf("%v", err)
			}

			pictureData.mark = string(pageDataBuffer[0:4])
			pictureData.width = binary.LittleEndian.Uint16(pageDataBuffer[4:6])
			pictureData.height = binary.LittleEndian.Uint16(pageDataBuffer[6:8])
			pictureData.colorMode = pageDataBuffer[8]
			pictureData.compression = pageDataBuffer[9]
			pictureData.dataSize = binary.LittleEndian.Uint32(pageDataBuffer[10:14])
			pictureData.md5 = binary.LittleEndian.Uint64(pageDataBuffer[14:22])
			pictureData.Data = [96000]byte(pageDataBuffer[22:])

			pictures = append(pictures, pictureData)
		}

		return pictures, nil
	} else {
		return nil, fmt.Errorf("test")
	}
}

//given path will return a slice of bytes
func GetXTGData(path string, buf []byte) int {
	fmt.Println(path)
	filePtr, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	fmt.Println(filePtr)

	i, err :=filePtr.ReadAt(buf, 22)	
	if err != nil {
		panic(err)
	}
	return i
}

func ExpandBitmap(data []byte) []byte{
	var tempData []byte
	for _, value := range data {
		stringy := fmt.Sprintf("%08b",value)
		for _, v := range stringy {
			switch v {
			case '1':
				tempData = append(tempData, 0xFF)
			case '0':
				tempData = append(tempData, 0x00)
			default:
				fmt.Println("unknown")
			}	
		}

	}	

	return tempData
}

func ExpandXTHBitmap(data []byte) []byte {
	var bPlanes []byte
		inner := 0
		outer := 48000
		if inner != 48000 && outer != 96000 {
			firstString := fmt.Sprintf("%08b",data[inner])
			secondString := fmt.Sprintf("%08b", data[outer])
			
			for i:=0; i!=8; i++ {
				tempString := string(firstString[i]) + string(secondString[i])
				switch tempString {
				case "11":
					bPlanes = append(bPlanes, 0xFF)

				case "01":
					bPlanes = append(bPlanes, 0xC0)

				case "10":
					bPlanes = append(bPlanes, 0x40)

				case "00":
					bPlanes = append(bPlanes, 0x00)
				
				default:
				fmt.Println("unknown")	
				}
			}
			inner++
			outer++
		}

	return bPlanes
}