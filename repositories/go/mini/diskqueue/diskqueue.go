package diskqueue

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"

	stdio "io"

	"github.com/NoFacePeace/github/repositories/go/utils/file"
	"github.com/NoFacePeace/github/repositories/go/utils/io"
)

type Interface interface {
	// 写入数据
	Put([]byte) error
	// 读取数据
	ReadChan() <-chan []byte
	// 读取数据，但不提交
	PeakChan() <-chan []byte
	Close() error
	Delete() error
	Depth() int64
	Empty() error
}

var (
	ErrExit = errors.New("exiting")
)

type diskQueue struct {
	// 读位置
	readPos int64
	// 写位置
	writePos int64
	// 读文件序号
	readFileNum int64
	// 写文件序号
	writeFileNum int64
	// 未消费总数
	depth int64

	// 读写锁，防止退出的时候还在写入，不是用于写入的时候加锁
	sync.RWMutex

	// 退出标志位，1 为退出
	exitFlag int32

	// 写入 channel
	writeChan chan []byte
	// 写入返回 channel
	writeResponseChan chan error

	// 读 channel
	readChan chan []byte

	// 读不提交 channel
	peakChan chan []byte

	// 退出 channel
	// 退出时关闭
	exitChan chan int

	// 退出同步 channel
	exitSyncChan chan int

	// 深度 channel
	depthChan chan int64

	// 写文件
	writeFile *os.File

	// 读文件
	readFile *os.File

	// 是否需要同步数据
	needSync bool

	// 文件路径
	dataPath string

	// 文件名
	name string

	// 是否为空 channel
	emptyChan chan int
	// 是否为空响应 channel
	emptyResponseChan chan error
	// 文件最大值
	maxBytesPerFile int64
	// 消息最小值
	minMsgSize int32
	// 消息最大值
	maxMsgSize int32
	// 同步间隔
	syncEvery int
	// 同步超时
	syncTimeout time.Duration
	//
	nextReadFileNum int64
	nextReadPos     int64
	// 读文件大小
	maxBytesPerFileRead int64

	// 读缓冲
	reader *bufio.Reader

	// 写缓冲
	writeBuf bytes.Buffer
}

func New(name string, path string, maxBytesPerFile int64, minMsgSize int32, maxMsgSize int32, syncEvery int, syncTimeout time.Duration) Interface {
	d := &diskQueue{
		name:              name,
		dataPath:          path,
		maxBytesPerFile:   maxBytesPerFile,
		minMsgSize:        minMsgSize,
		maxMsgSize:        maxMsgSize,
		readChan:          make(chan []byte),
		peakChan:          make(chan []byte),
		depthChan:         make(chan int64),
		writeChan:         make(chan []byte),
		writeResponseChan: make(chan error),
		emptyChan:         make(chan int),
		emptyResponseChan: make(chan error),
		exitChan:          make(chan int),
		exitSyncChan:      make(chan int),
		syncEvery:         syncEvery,
		syncTimeout:       syncTimeout,
	}
	if err := d.retrieveMetaData(); err != nil && !os.IsNotExist(err) {
		slog.Error("disk queue retrieve meta data error", "error", err)
	}
	go d.ioLoop()
	return d
}

// 写入数据, 支持并发
func (d *diskQueue) Put(data []byte) error {
	// 写入数据的时候加读锁，防止退出的时候还在写入数据
	// 退出的时候使用的是写锁
	// 写入数据不需要加写锁，使用的是 chan，不会有并发问题
	d.RLock()
	defer d.RUnlock()
	if d.exitFlag == 1 {
		return ErrExit
	}
	// 写入 chan, 等待响应 chan 返回
	d.writeChan <- data
	return <-d.writeResponseChan
}

func (d *diskQueue) ReadChan() <-chan []byte {
	return d.readChan
}

func (d *diskQueue) PeakChan() <-chan []byte {
	return d.peakChan
}

func (d *diskQueue) Close() error {
	err := d.exit(false)
	if err != nil {
		return err
	}
	return d.sync()
}

func (d *diskQueue) exit(deleted bool) error {
	// 加锁，禁止继续写入数据
	d.Lock()
	defer d.Unlock()

	// 设置删除标志位
	d.exitFlag = 1

	if deleted {
		slog.Info("delete")
	}

	// 关闭退出 channel
	close(d.exitChan)

	// 等待死循环结束
	<-d.exitSyncChan

	// 关闭深度 channel
	close(d.depthChan)
	if d.writeFile != nil {
		d.writeFile.Close()
		d.writeFile = nil
	}
	if d.readFile != nil {
		d.readFile.Close()
		d.readFile = nil
	}
	return nil
}

// 同步数据到磁盘, 并持久化元数据
func (d *diskQueue) sync() error {
	if d.writeFile != nil {
		err := d.writeFile.Sync()
		if err != nil {
			d.writeFile.Close()
			d.writeFile = nil
			return err
		}
	}
	err := d.persistMetaData()
	if err != nil {
		return err
	}
	d.needSync = false
	return nil
}

func (d *diskQueue) persistMetaData() error {
	fileName := d.metaDataFileName()
	data := fmt.Sprintf("%d\n%d,%d\n%d,%d\n", d.depth, d.readFileNum, d.readPos, d.writeFileNum, d.writePos)
	return file.PersistMetaData(fileName, []byte(data))
}

func (d *diskQueue) metaDataFileName() string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.meta.dat"), d.name)
}

func (d *diskQueue) Delete() error {
	return d.exit(true)
}

func (d *diskQueue) Depth() int64 {
	depth, ok := <-d.depthChan
	if !ok {
		// 循环退出
		depth = d.depth
	}
	return depth
}

func (d *diskQueue) Empty() error {
	d.RLock()
	defer d.RUnlock()
	if d.exitFlag == 1 {
		return ErrExit
	}
	d.emptyChan <- 1
	return <-d.emptyResponseChan
}

func (d *diskQueue) retrieveMetaData() error {
	var f *os.File
	var err error
	fileName := d.metaDataFileName()
	f, err = os.OpenFile(fileName, os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("os open file error: [%w]", err)
	}
	defer func() {
		err = io.SafeClose(f, err)
	}()
	var depth int64
	if _, err := fmt.Fscanf(f, "%d\n%d,%d\n%d,%d\n", depth, &d.readFileNum, &d.readPos, &d.writeFileNum, &d.writePos); err != nil {
		return fmt.Errorf("fmt fscanf error: [%w]", err)
	}
	d.depth = depth
	d.nextReadFileNum = d.readFileNum
	d.nextReadPos = d.readPos

	// 如果文件大小大于读位置，meta data 同步报错，需要新建文件
	fileName = d.fileName(d.writeFileNum)
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return fmt.Errorf("os stat error: [%w]", err)
	}
	fileSize := fileInfo.Size()
	if d.writePos < fileSize {
		slog.Warn("disk queue meta date write position < file size, skipping to new file")
		d.writeFileNum += 1
		d.writePos = 0
		if d.writeFile != nil {
			if err := d.writeFile.Close(); err != nil {
				return fmt.Errorf("disk queue write file close error: [%w]", err)
			}
			d.writeFile = nil
		}
	}
	return nil
}

func (d *diskQueue) ioLoop() {
	var r chan []byte
	var p chan []byte
	var count int
	var err error
	var dataRead []byte
	syncTicker := time.NewTicker(d.syncTimeout)
	// 协程，死循环
	for {
		// 循环次数，判断是否需要同步
		if count == d.syncEvery {
			d.needSync = true
		}

		// 需要同步，重置循环次数
		if d.needSync {
			if err := d.sync(); err != nil {
				slog.Error("disk queue sync error", "error", err)
			}
			count = 0
		}

		// 判断是否需要读一条数据
		if d.readFileNum < d.writeFileNum || d.readPos < d.writePos {
			if d.nextReadPos == d.readPos {
				dataRead, err = d.readOne()
				if err != nil {
					slog.Error("disk queue read one error", "error", err)
					// 处理读异常
					d.handleReadError()
					continue
				}
			}
			r = d.readChan
			p = d.peakChan
		} else {
			r = nil
			p = nil
		}
		select {
		case p <- dataRead:
		case r <- dataRead:
			count++
			d.moveForward()
		case <-d.emptyChan:
			d.emptyResponseChan <- d.deleteAllFiles()
			count = 0
		case dataWrite := <-d.writeChan:
			count++
			d.writeResponseChan <- d.writeOne(dataWrite)
		case <-syncTicker.C:
			if count == 0 {
				continue
			}
			d.needSync = true
		case <-d.exitChan:
			goto exit
		}
	}
exit:
	slog.Info("disk queue close io loop")
	syncTicker.Stop()
	d.exitSyncChan <- 1
}

func (d *diskQueue) fileName(fileNum int64) string {
	return fmt.Sprintf(path.Join(d.dataPath, "%s.diskqueue.%06d.dat"), d.name, fileNum)
}

func (d *diskQueue) readOne() ([]byte, error) {
	var err error
	var msgSize int32

	// 如果读文件为空，则打开文件
	if d.readFile == nil {
		curFileName := d.fileName(d.readFileNum)
		d.readFile, err = os.OpenFile(curFileName, os.O_RDONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("os open file error: [%w]", err)
		}

		// 重新定位到读位置
		if d.readPos > 0 {
			if _, err = d.readFile.Seek(d.readPos, 0); err != nil {
				d.readFile.Close()
				d.readFile = nil
				return nil, fmt.Errorf("disk queue read file seek error: [%w]", err)
			}

		}
		d.maxBytesPerFileRead = d.maxBytesPerFile
		if d.readFileNum < d.writeFileNum {
			stat, err := d.readFile.Stat()
			if err != nil {
				return nil, fmt.Errorf("disk queue read file stat error: [%w]", err)
			}
			d.maxBytesPerFileRead = stat.Size()
		}
		d.reader = bufio.NewReader(d.readFile)
	}
	// 读下一个信息大小
	if err := binary.Read(d.reader, binary.BigEndian, &msgSize); err != nil {
		d.readFile.Close()
		d.readFile = nil
		return nil, fmt.Errorf("binary read error: [%w]", err)
	}
	if msgSize < d.minMsgSize || msgSize > d.maxMsgSize {
		d.readFile.Close()
		d.readFile = nil
		return nil, fmt.Errorf("invalid message read size (%d)", msgSize)
	}
	// 读取消息
	readBuf := make([]byte, msgSize)
	if _, err := stdio.ReadFull(d.reader, readBuf); err != nil {
		d.readFile.Close()
		d.readFile = nil
		return nil, fmt.Errorf("io read full error: [%w]", err)
	}
	totalBytes := int64(4 + msgSize)
	d.nextReadPos = totalBytes + d.readPos
	d.nextReadFileNum = d.readFileNum

	// 如果下一个读位置大于最大的文件大小，则切换到下一个文件
	if d.readFileNum < d.writeFileNum && d.nextReadPos >= d.maxBytesPerFileRead {
		if d.readFile != nil {
			d.readFile.Close()
			d.readFile = nil
		}
		d.nextReadFileNum++
		d.nextReadPos = 0
	}
	return readBuf, nil
}

func (d *diskQueue) handleReadError() {
	// 如果读文件和写文件一致，关闭写文件，新建文件
	if d.readFileNum == d.writeFileNum {
		if d.writeFile != nil {
			d.writeFile.Close()
			d.writeFile = nil
		}
		d.writeFileNum++
		d.writePos = 0
	}

	// 将异常文件备份
	badFn := d.fileName(d.readFileNum)
	badRenameFn := badFn + ".bad"
	if err := os.Rename(badFn, badRenameFn); err != nil {
		slog.Error("os rename error", "error", err)
	}

	// 重置读文件元数据
	d.readFileNum++
	d.readPos = 0
	d.nextReadFileNum = d.readFileNum
	d.nextReadPos = 0
	d.needSync = true

	// 检查读写元数据是否异常k m
	d.checkTailCorruption(d.depth)
}

func (d *diskQueue) moveForward() {
	oldReadFileNum := d.readFileNum
	d.readFileNum = d.nextReadFileNum
	d.readPos = d.nextReadPos
	d.depth -= 1
	if oldReadFileNum != d.nextReadFileNum {
		d.needSync = true
		fn := d.fileName(oldReadFileNum)
		if err := os.Remove(fn); err != nil {
			slog.Error("os remove", "error", err)
		}
	}
	d.checkTailCorruption(d.depth)
}

func (d *diskQueue) deleteAllFiles() error {
	if err := d.skipToNextRWFile(); err != nil {
		return fmt.Errorf("diskqueue skip to next rw file error: [%w]", err)
	}
	err := os.Remove(d.metaDataFileName())
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("os remove error: [%w]", err)
	}
	return nil
}

func (d *diskQueue) writeOne(data []byte) (err error) {
	dataLen := int32(len(data))
	totalBytes := int64(4 + dataLen)
	if dataLen < d.minMsgSize || dataLen > d.maxMsgSize {
		return fmt.Errorf("invalid message write size (%d) minMsgSize=%d maxMsgSize=%d", dataLen, d.minMsgSize, d.maxMsgSize)
	}

	// 检查是否超过文件大小
	// 超过文件大小，重新新建文件
	if d.writePos > 0 && d.writePos+totalBytes > d.maxBytesPerFile {
		if d.readFileNum == d.writeFileNum {
			d.maxBytesPerFileRead = d.writePos
		}
		d.writeFileNum++
		d.writePos = 0
		if err := d.sync(); err != nil {
			slog.Error("disk queue sync error: [%w]", "error", err)
		}
		if d.writeFile != nil {
			d.writeFile.Close()
			d.writeFile = nil
		}
	}

	// 检查写文件是否打开
	if d.writeFile == nil {
		curFileName := d.fileName(d.writeFileNum)
		d.writeFile, err = os.OpenFile(curFileName, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		if d.writePos != 0 {
			if _, err = d.writeFile.Seek(d.writePos, 0); err != nil {
				d.writeFile.Close()
				d.writeFile = nil
				return err
			}
		}
	}
	d.writeBuf.Reset()
	if err := binary.Write(&d.writeBuf, binary.BigEndian, dataLen); err != nil {
		return err
	}
	if _, err := d.writeBuf.Write(data); err != nil {
		return err
	}
	if _, err := d.writeFile.Write(d.writeBuf.Bytes()); err != nil {
		d.writeFile.Close()
		d.writeFile = nil
		return err
	}
	d.writePos += totalBytes
	d.depth += 1
	return nil
}

// 检查读写元数据是否异常
func (d *diskQueue) checkTailCorruption(depth int64) {
	if d.readFileNum < d.writeFileNum || d.readPos < d.writePos {
		return
	}
	if depth != 0 {
		d.depth = 0
		d.needSync = true
	}
	if d.readFileNum != d.writeFileNum || d.readPos != d.writePos {
		d.skipToNextRWFile()
		d.needSync = true
	}
}

// 跳过损坏文件
func (d *diskQueue) skipToNextRWFile() (err error) {
	// 关闭读文件
	if d.readFile != nil {
		d.readFile.Close()
		d.readFile = nil
	}

	// 关闭写文件
	if d.writeFile != nil {
		d.writeFile.Close()
		d.writeFile = nil
	}

	// 删除损坏文件
	for i := d.readFileNum; i <= d.writeFileNum; i++ {
		fn := d.fileName(i)
		innerErr := os.Remove(fn)
		if innerErr != nil && !os.IsNotExist(innerErr) {
			err = innerErr
		}
	}

	// 重置读写文件元数据
	d.writeFileNum++
	d.writePos = 0
	d.readFileNum = d.writeFileNum
	d.readPos = 0
	d.nextReadFileNum = d.writeFileNum
	d.nextReadPos = 0
	d.depth = 0
	return
}
