package oss

import (
	"context"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"time"
)

const (
	BUCKET_NAME = "sunjinyan-testing"
)

type Service struct {
	client *oss.Client
	accessKeyID string
	accessKeySecret string
}

/**
accessKeyID := "LTAI5tNkUusc4tqY73p6k4YZ"
accessKeySecret := "6JHZEAKZAphr5OzAjzNam2jJRXaa7M"
endpoint	:= "oss-cn-beijing.aliyuncs.com"
 */

func NewService(endpoint,accessKeyID,accessKeySecret string) (*Service , error) {

	// 创建OSSClient实例。
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		//fmt.Println("Error:", err)
		//os.Exit(-1)
		return nil, fmt.Errorf("cannot create oss client:%v",err)
	}

	return &Service{
		client: client,
		accessKeyID: accessKeyID,
		accessKeySecret: accessKeySecret,
	},err
}


//func (s *Service)SignURL(c context.Context,,bucket string,timeout time.Duration) (string,error){
//
//}


func (s *Service)Get(c context.Context,path string)(io.ReadCloser,error){
	// 获取存储空间。
	bucket, err := s.client.Bucket(BUCKET_NAME)
	if err != nil {
		//fmt.Println("Error:", err)
		//os.Exit(-1)
		return nil, fmt.Errorf("cannot connect Bucket , error info : %v",err)
	}

	// 下载文件到流。
	body, err := bucket.GetObject(path)
	if err != nil {
		//fmt.Println("Error:", err)
		//os.Exit(-1)
		return nil, fmt.Errorf("cannot Get  Object, error info : %v",err)
	}
	return body, nil
}

func (s *Service)SignURL(c context.Context,method,path string,timeout time.Duration) (string,error){
	//func (s *Service) SignURL(AccessKeyId string, AccessKeySecret string, path string, operation int64) (string, error) {
	credentials, err := s.GetCredentials(s.accessKeyID, s.accessKeySecret)
	if err != nil {
		return "", fmt.Errorf("cannot create GetCredentials , error info : %v",err)
	}

	/**
	  创建客户端
	  客户端需一个节点信息，这里选择了杭州阿里云节点
	  需要的临时key，secret，token从 sts.Credentials 中获取
	*/
	client, err := oss.New("http://oss-cn-beijing.aliyuncs.com", credentials.AccessKeyId, credentials.AccessKeySecret, oss.SecurityToken(credentials.SecurityToken))
	if err != nil {
		return "", fmt.Errorf("cannot create client , error info : %v",err)
	}

	bucketName := BUCKET_NAME
	bucket, err := client.Bucket(bucketName)

	if err != nil {
		return "", fmt.Errorf("cannot connect Bucket , error info : %v", err)
	}

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
	if method == "PUT" {
		operationAction = oss.HTTPPut
	}


	signedURL, err := bucket.SignURL(path, operationAction,3600, options...)
	if err != nil {
		return "", fmt.Errorf("cannot connect SignURL , error info : %v", err)
	}

	//去除http请求的临时图片地址

	err = bucket.PutObjectFromFileWithURL(signedURL, "D:/coolcar/wx/miniprogram/resources/lic.png", options...)
	if err != nil {
		return "", fmt.Errorf("cannot connect SignURL , error info : %v", err)
	}
	return signedURL, nil
}

func (s *Service) GetCredentials(AccessKeyId string, AccessKeySecret string) (*sts.Credentials, error) {
	client, err := sts.NewClientWithAccessKey("cn-beijing", AccessKeyId, AccessKeySecret)

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
	request.RoleArn = "acs:ram::1489471921297128:role/osstesting"
	request.RoleSessionName = "osstesting"

	response, err := client.AssumeRole(request)
	if err != nil {
		return nil, fmt.Errorf("cannot create AssumeRole , error info : %v",err)
	}

	/**
	  返回临时身份信息
	*/
	return &response.Credentials, nil
}




func (s *Service)createBucket(bucketName string) (err error) {
	if bucketName == "" {
		return fmt.Errorf("bucket name can not by empty")
		//bucketName = BUCKET_NAME
	}
	// 创建名为examplebucket的存储空间，并设置存储类型为低频访问oss.StorageIA、读写权限ACL为公共读oss.ACLPublicRead、数据容灾类型为同城冗余存储oss.RedundancyZRS。
	err = s.client.CreateBucket(bucketName, oss.StorageClass(oss.StorageIA), oss.ACL(oss.ACLPublicRead), oss.RedundancyType(oss.RedundancyZRS))
	if err != nil {
		//fmt.Println("Error:", err)
		//os.Exit(-1)
		return fmt.Errorf("cannot create bucket , error info : %v",err)
	}
	return err
}



















