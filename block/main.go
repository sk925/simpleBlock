package main

import (
	"os"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"io"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
)
type Block struct {
	Index int
	TimeStamp string
	Data string
	PreHash string
	CurrentHash string
}
var BlockChain = make([]Block, 0)
//存放请求的信道满足高并发下区块的创建和添加
var dataChannel = make(chan string,5)
//计算区块的哈希
func CalculateHash(block Block) string{
	str := string(block.Index) + block.TimeStamp + block.Data + block.PreHash
	hash := sha256.New()
	hash.Write([]byte(str))
	hashCode := hash.Sum(nil)
	return hex.EncodeToString(hashCode)
}
//判断是不是有效的区块
func CreateBlock(preBlock Block, data string) Block{
	block := Block{}
	block.Index =preBlock.Index + 1
	block.Data = data
	block.PreHash = preBlock.CurrentHash
	block.TimeStamp = time.Now().String()
	block.CurrentHash = CalculateHash(block)
	return block
}

func IsValid(block Block, preBlock Block) bool{
	if preBlock.Index + 1 != block.Index{
		return false
	}
	if preBlock.CurrentHash != block.PreHash {
		return false
	}
	if CalculateHash(block) != block.CurrentHash {
		return false
	}
	return true
}

func main() {
	err := godotenv.Load("server.env")
	if err != nil {
		log.Fatal(err)
	}
	go func(){
		first := Block{0,time.Now().String(),"","",""}
		first.CurrentHash = CalculateHash(first)
		spew.Dump(first)
		BlockChain = append(BlockChain,first)
	}()
	go addBlock()
	log.Fatal(run())
}

func run() error{
	handler := makeHandler()
	port := os.Getenv("port")
	log.Println("server listening on", port)
	server := &http.Server{
		Addr: ":" + port,
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := server.ListenAndServe(); err != nil{
		return err
	}
	return nil
}

func makeHandler() http.Handler{
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/get",getBlockChain).Methods("Get")
	muxRouter.HandleFunc("/add",receiveData).Methods("Get")
	return muxRouter
}

func getBlockChain(responseWriter http.ResponseWriter, request *http.Request){
	bytes,err := json.MarshalIndent(BlockChain,""," ")
	if err != nil {
		http.Error(responseWriter,err.Error(),http.StatusInternalServerError)
		return
	}
	io.WriteString(responseWriter,string(bytes))
}

func receiveData(responseWriter http.ResponseWriter, request *http.Request){
	data := request.FormValue("data")
	dataChannel <- data
	spew.Dump(data)
	io.WriteString(responseWriter,"data has send to channel")
}
func addBlock(){
	for data := range dataChannel{
		preBlock := BlockChain[len(BlockChain) - 1]
		newBlock := CreateBlock(preBlock,data)
		spew.Dump(newBlock)
		if IsValid(newBlock,preBlock) {
			BlockChain = append(BlockChain,newBlock)
			spew.Dump(BlockChain)
		}
	}
}



