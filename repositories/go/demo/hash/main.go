package main

import (
	"fmt"
	"hash/crc32"
)

func main() {
	hasher := crc32.NewIEEE()
	hasher.Write([]byte("persistent:/public/probe/topic-partition-0"))
	num := hasher.Sum32()
	fmt.Printf("%X\n", num)
}

// func main() {
// 	topic := "smq-test-kafka-topic-1"

// 	// 使用 IEEE 算法计算 CRC32 哈希值
// 	hasher := crc32.NewIEEE()
// 	hasher.Write([]byte(topic))
// 	hash := hasher.Sum32()

//		fmt.Printf("CRC32 哈希值: %d (十进制)\n", hash)
//		fmt.Printf("CRC32 哈希值: 0x%X (十六进制)\n", int(hash))
//	}
// func main() {
// 	topic := "smq-test-kafka-topic-partition-0"

// 	// 使用 IEEE 算法
// 	ieeeTable := crc32.MakeTable(crc32.IEEE)
// 	ieeeHash := crc32.Checksum([]byte(topic), ieeeTable)

// 	// 使用 Castagnoli 算法
// 	castagnoliTable := crc32.MakeTable(crc32.Castagnoli)
// 	castagnoliHash := crc32.Checksum([]byte(topic), castagnoliTable)

// 	fmt.Printf("CRC32 (IEEE): %d\n", ieeeHash)
// 	fmt.Printf("CRC32 (Castagnoli): %d\n", castagnoliHash)
// }
