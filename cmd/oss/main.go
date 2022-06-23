package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"net/http"
	"os"
)

/*
LTAI5tM4UPPdcKhoYqbnmJGx
5cJfdeVio9Ih1vYimdIY2soLUrcNeh
*/
func main() {
	client, err := oss.New("oss-cn-beijing.aliyuncs.com", "LTAI5tNkUusc4tqY73p6k4YZ", "6JHZEAKZAphr5OzAjzNam2jJRXaa7M")

	if err != nil {
		fmt.Printf("create new oss client error , error info:%+v",err)
		return
	}
	lsRes, err := client.ListBuckets()
	if err != nil {
		fmt.Printf("list buckets error , error info:%+v",err)
		// HandleError(err)
		return
	}

	for _, bucket := range lsRes.Buckets {
		fmt.Println("Buckets:", bucket.Name)
	}

	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	//
	//go func() {
	//	defer wg.Done()
		//UploadObjectFromLocalToOss()
	//}()
	//wg.Wait()
	GetObjectFromStream()
}

func UploadObjectFromLocalToOss() {
	client, err := oss.New("oss-cn-beijing.aliyuncs.com", "LTAI5tM4UPPdcKhoYqbnmJGx", "5cJfdeVio9Ih1vYimdIY2soLUrcNeh")

	if err != nil {
		fmt.Printf("create new oss client error , error info:%+v",err)
		return
	}

	bucket, err := client.Bucket("sunjinyan-testing")
	if err != nil {
		panic(err)
	}

	err = bucket.PutObjectFromFile("my-object", "D:/coolcar/wx/miniprogram/resources/lic.png")
	if err != nil {
		panic(err)
	}
}

func GetObject() {
	client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
	if err != nil {
		// HandleError(err)
	}

	bucket, err := client.Bucket("my-bucket")
	if err != nil {
		// HandleError(err)
	}

	err = bucket.GetObjectToFile("my-object", "LocalFile")
	if err != nil {
		// HandleError(err)
	}
}

func GetObjectFromStream() {
	// 创建OSSClient实例。
	client, err := oss.New("oss-cn-beijing.aliyuncs.com", "LTAI5tNkUusc4tqY73p6k4YZ", "6JHZEAKZAphr5OzAjzNam2jJRXaa7M")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 获取存储空间。
	bucket, err := client.Bucket("sunjinyan-testing")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 下载文件到流。
	body, err := bucket.GetObject("my-object")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄漏，导致请求无连接可用，程序无法正常工作。
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	buf := bytes.NewBuffer(data)

	var base64Encoding string

	// 判断文件类型，生成一个前缀，拼接base64后可以直接粘贴到浏览器打开，不需要可以不用下面代码
	//取图片类型
	mimeType := http.DetectContentType(data)
	switch mimeType {
	case "image/jpeg":
		base64Encoding = "data:image/jpeg;base64," + base64Encoding
	case "image/png":
		base64Encoding = "data:image/png;base64," + base64Encoding
	}
	base64Encoding += base64.StdEncoding.EncodeToString(buf.Bytes())//生成base64,使用图片base64解码，就可以查看到原图，开始我不懂，直接用base64解码，结果尴尬了
	/*fileHandle, err1 := os.OpenFile("image.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666);

	if err1 != nil {
		log.Println("open file err1or :", err1)
		return
	}

	// NewWriter 默认缓冲区大小是 4096
	// 需要使用自定义缓冲区的writer 使用 NewWriterSize()方法
	buf1 := bufio.NewWriterSize(fileHandle, len(base64Encoding))

	buf1.WriteString(base64Encoding)

	buf1.Flush()
	fileHandle.Close()*/


	fmt.Println("data:", base64Encoding)
	//fmt.Println("data:", string(data))
}

type OssService struct {
}

func (o *OssService) CreateSingUrl(AccessKeyId string, AccessKeySecret string, path string, operation int64) (string, error) {
	credentials, err := o.GetCredentials(AccessKeyId, AccessKeySecret)
	if err != nil {
		return "", err
	}

	/**
	  创建客户端
	  客户端需一个节点信息，这里选择了杭州阿里云节点
	  需要的临时key，secret，token从 sts.Credentials 中获取
	*/
	client, err := oss.New("http://oss-cn-hangzhou.aliyuncs.com", credentials.AccessKeyId, credentials.AccessKeySecret, oss.SecurityToken(credentials.SecurityToken))
	if err != nil {
		return "", err
	}

	bucketName := "study-golang"
	bucket, err := client.Bucket(bucketName)

	/**
	  这里以png图片为例，故此设置为 image/png
	*/
	options := []oss.Option{
		oss.ContentType("image/png"),
	}

	/**
	获取预签名url
	指定了该url存储位置，提交方式，有效时间，附加参数
	oss.HTTPGet 代表生成的url可以用来下载
	oss.HTTPPut 代表生成的url可以用来上传
	*/
	operationAction := oss.HTTPGet
	if operation == 2 {
		operationAction = oss.HTTPPut
	}


	signedURL, err := bucket.SignURL(path, operationAction, 6000, options...)

	return signedURL, err
}

func (o *OssService) GetCredentials(AccessKeyId string, AccessKeySecret string) (*sts.Credentials, error) {
	client, err := sts.NewClientWithAccessKey("cn-hangzhou", AccessKeyId, AccessKeySecret)

	if err != nil {
		return nil, err
	}

	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"

	/**
	  访问 https://ram.console.aliyun.com/roles 可以看到
	  要保证该角色有权限操作oss
	  RoleArn 即 该角色的Arn
	  RoleSessionName 标识名称
	*/
	request.RoleArn = "acs:ram::xxxxxxxxxx:role/osstest"
	request.RoleSessionName = "ossTest"

	response, err := client.AssumeRole(request)
	if err != nil {
		return nil, err
	}

	/**
	  返回临时身份信息
	*/
	return &response.Credentials, nil
}