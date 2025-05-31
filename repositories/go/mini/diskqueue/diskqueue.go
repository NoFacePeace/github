package diskqueue

import (
	"errors"
	"os"
	"sync"
)

type Interface interface {
	// 写入数据
	Put([]byte) error
	// 读取数据
	ReadChan() <-chan []byte
	// 读取数据，但不提交
	PeakChan() <-chan []byte
	Close() error
	// Delete() error
	// Depth() error
	// Empty() error
}
type diskQueue struct {
	// 读位置
	readPos int64
	// 写位置
	writePos int64
	// 读文件序号
	readFileNum int64
	// 写文件序号
	writeFileNum int64
	// 待定
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
}

func New(name string) Interface {
	q := &diskQueue{}
	return q
}

// 写入数据, 支持并发
func (d *diskQueue) Put(data []byte) error {
	// 写入数据的时候加读锁，防止退出的时候还在写入数据
	// 退出的时候使用的是写锁
	// 写入数据不需要加写锁，使用的是 chan，不会有并发问题
	d.RLock()
	defer d.RUnlock()
	if d.exitFlag == 1 {
		return errors.New("exiting")
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
	return nil
}
