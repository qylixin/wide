package remote

import (
	"encoding/json"
	"fmt"
	"github.com/b3log/wide/conf"
	"time"
)

type QueryResInfo struct {
	PageSum int         `json:"page_sum"`
	Data    interface{} `json:"data"`
}

type DetailChannelInfo struct {
	*Channel
	OrgCount       int          `json:"org_count"`       // 通道关联组织数
	ChaincodeCount int          `json:"chaincode_count"` // 通道关联合约数
	State          int          `json:"state"`           //0未验证，1等待中，2成功，3失败，4网络连接失败
	Ccs            []*Chaincode `json:"ccs"`             //通道对应的合约
}

type DetailCcInfo struct {
	*Chaincode
	BindOrgCount int    `json:"bind_org_count"` //绑定组织个数
	ChannelName  string `json:"channel_name"`   //通道名称
	ChannelType  int    `json:"channel_type"`   //通道类型
	StartValid   bool   `json:"start_valid"`    //是否能够启动
	UpgradeValid bool   `json:"upgrade_valid"`  //是否能够升级
}

type CCBase struct {
	ID                  uint      `json:"id" gorm:"primary_key;AUTO_INCREMENT;column:id" form:"id"` //需要做唯一索引,所以必须存在。
	UUID                string    `json:"uuid" gorm:"not null,index"`                               //后端识别名
	Version             string    `json:"version"`                                                  //版本
	NetUUID             string    `json:"net_uuid" gorm:"not null,index"`                           //对应网络
	ChannelUUID         string    `json:"channel_uuid"`                                             //对应channel
	EndorsePolicy       string    `json:"endorse_policy" gorm:"type:text"`                          //背书策略
	InitParam           string    `json:"init_param" gorm:"type:text"`                              //初始化参数
	PkgInfo             string    `json:"pkg_info"`                                                 //上传文件hash
	ChaincodePath       string    `json:"chaincode_path"`                                           //合约路径
	ChaincodeFileName   string    `json:"chaincode_file_name"`                                      //合约的名称
	State               int       `json:"state"`                                                    //状态（0: 未启动， 1:异常, 2：启动成功, 3：升级未启动 4.待升级）
	InstState           int       `json:"inst_state"`                                               //实例化状态(0:未实例化，1:已经实例化过了)
	PeerUUIDs           string    `json:"peer_uuids"`                                               //已安装cc的peer列表
	PrePeerUUIDs        string    `json:"pre_peer_uuids"`                                           //预安装cc的peer列表
	InstPeerUUIDs       string    `json:"inst_peer_uuids"`                                          //已实例化的Peer列表
	InstUnBindPeerUUIDs string    `json:"inst_un_bind_peer_uuids"`                                  //已实例化但是已解绑的Peer列表
	OrgUUIDs            string    `json:"-"`                                                        //实例化前绑定通道的组织列表
	Type                string    `json:"type"`                                                     //安装方式 0：源码上传， 1：在线合约编辑
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Chaincode struct {
	CCBase
	Name string `gorm:"not null" json:"name"` //名称(任意字符串)
}

type Channel struct {
	ID        uint      `json:"id" gorm:"primary_key"`           //需要做唯一索引,所以必须存在。
	UUID      string    `json:"uuid" gorm:"not null,index"`      //后端识别名
	Name      string    `gorm:"not null" json:"name"`            //名称(任意字符串)
	NetUUID   string    `json:"net_uuid"  gorm:"not null,index"` //对应网络
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      uint      `json:"type" gorm:"not null"` //channel类型,0共有，1私有，2需验证
}

func GetChannel(netuuid, user, token string) ([]*DetailChannelInfo, error) {

	fmt.Println("address : ", conf.Wide.BcapAddress)
	request := NewRequest(conf.Wide.BcapAddress, "/chaincode/channels")
	request.SetQuery("net_uuid", netuuid)
	request.SetHeader("user", user)
	request.SetHeader("token", token)
	res := request.GET()
	if res.Error != nil {
		return nil, res.Error
	}

	response := &ResponseInfo{}
	data := make([]*DetailChannelInfo, 0)
	response.Data = &data

	fmt.Println("=========== ", string(res.Data))

	if err := json.Unmarshal(res.Data, response); err != nil {
		return nil, err
	}

	return data, nil
}

func InstallChaincode(netuuid, channeluuid, path, name, user, token string) (*ResponseInfo, error) {

	fmt.Println("address : ", conf.Wide.BcapAddress)

	request := NewRequest(conf.Wide.BcapAddress, "/chaincode")
	request.SetQuery("net_uuid", netuuid)
	request.SetQuery("channel_uuid", channeluuid)
	request.SetQuery("type", "1")
	request.SetQuery("name", name)
	request.SetHeader("user", user)
	request.SetHeader("token", token)
	request.SetFile(path)

	res := request.POST()

	if res.Error != nil {
		return nil, res.Error
	}

	response := &ResponseInfo{}
	data := &Chaincode{}
	response.Data = &data

	if err := json.Unmarshal(res.Data, response); err != nil {
		return nil, err
	}

	fmt.Println("******************** install  chaincode  *********************")
	fmt.Println(response)
	fmt.Println("******************** install chaincode  *********************")

	return response, nil
}

func UpgradeChaincode(netuuid, ccid, path, user, token string) (*ResponseInfo, error) {

	fmt.Println("address : ", conf.Wide.BcapAddress)

	request := NewRequest(conf.Wide.BcapAddress, "/chaincode/"+ccid)
	request.SetQuery("net_uuid", netuuid)
	request.SetQuery("type", "1")
	request.SetHeader("user", user)
	request.SetHeader("token", token)
	request.SetFile(path)

	res := request.PUT()

	if res.Error != nil {
		return nil, res.Error
	}

	response := &ResponseInfo{}
	data := &Chaincode{}
	response.Data = &data

	if err := json.Unmarshal(res.Data, response); err != nil {
		return nil, err
	}

	fmt.Println("******************** upgrade chaincode  *********************")
	fmt.Println(data)
	fmt.Println("******************** upgrade chaincode  *********************")

	return response, nil
}

func GetChaincode(netuuid, user, token string) ([]*Chaincode, error) {

	fmt.Println("address : ", conf.Wide.BcapAddress)
	request := NewRequest(conf.Wide.BcapAddress, "/chaincodes")
	request.SetQuery("net_uuid", netuuid)
	request.SetQuery("page", "0")
	request.SetQuery("record_count", "0")
	request.SetQuery("sort_type", "0")
	request.SetQuery("is_asc", "1")
	request.SetHeader("user", user)
	request.SetHeader("token", token)
	res := request.GET()
	if res.Error != nil {
		return nil, res.Error
	}

	response := &ResponseInfo{}
	var queryResInfo QueryResInfo
	data := make([]*DetailCcInfo, 0)
	queryResInfo.Data = &data
	response.Data = &queryResInfo

	if err := json.Unmarshal(res.Data, response); err != nil {
		return nil, err
	}

	resData := make([]*Chaincode, 0)
	for _, dc := range data {
		resData = append(resData, dc.Chaincode)
	}

	fmt.Println("********************  chaincodes  *********************")
	fmt.Println(resData)
	fmt.Println("********************  chaincodes  *********************")

	return resData, nil
}
