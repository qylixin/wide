package remote

import (
	"encoding/json"
	"fmt"
	"github.com/b3log/wide/conf"
	"time"
)

type DetailChannelInfo struct {
	*Channel
	OrgCount int          `json:"orgCount"`
	State    int          `json:"state"` //0未验证，1等待中，2成功，3失败，4网络连接失败
	Ccs      []*Chaincode `json:"ccs"`   //通道对应的合约
}

type CCBase struct {
	ID            uint      `json:"id" gorm:"primary_key"`         //需要做唯一索引,所以必须存在。
	UUID          string    `json:"uuid" gorm:"not null,index"`    //后端识别名
	Version       string    `json:"version"`                       //版本
	NetUUID       string    `json:"netuuid" gorm:"not null,index"` //对应网络
	ChannelUUID   string    `json:"channeluuid"`                   //对应channel
	EndorsePolicy string    `json:"endorsePolicy"`                 //背书策略
	InitParam     string    `json:"initParam"`                     //初始化参数
	PkgInfo       string    `json:"pkgInfo"`                       //上传文件hash
	State         int       `json:"state"`                         //状态（0: 未启动， 1:异常, 2：启动成功, 3：升级未启动 4.待升级）
	InstState     int       `json:"inst_state"`                    //实例化状态(0:未实例化，1:已经实例化过了)
	PeerUUIDs     string    `json:"peeruuids"`                     //已安装cc的peer列表
	PrePeerUUIDs  string    `json:"prepeeruuids"`                  //预安装cc的peer列表
	InstPeerUUIDs string    `json:"instpeeruuids"`                 //已实例化的Peer列表
	OrgUUIDs      string    `json:"-"`                             //实例化前绑定通道的组织列表
	CreatedAt     time.Time `json:"createAt"`
	UpdatedAt     time.Time `json:"updateAt"`
}

type Chaincode struct {
	CCBase
	Name string `gorm:"not null" json:"name"` //名称(任意字符串)
}

type Channel struct {
	ID        uint      `json:"id" gorm:"primary_key"`          //需要做唯一索引,所以必须存在。
	UUID      string    `json:"uuid" gorm:"not null,index"`     //后端识别名
	Name      string    `gorm:"not null" json:"name"`           //名称(任意字符串)
	NetUUID   string    `json:"netuuid"  gorm:"not null,index"` //对应网络
	CreatedAt time.Time `json:"createAt"`
	UpdatedAt time.Time `json:"updateAt"`
	Type      uint      `json:"type" gorm:"not null"` //channel类型,0共有，1私有，2需验证
}

func GetChannel(netuuid, user, token string) ([]*DetailChannelInfo, error) {

	fmt.Println("address : ", conf.Wide.BcapAddress)
	request := NewRequest(conf.Wide.BcapAddress, "/channels")
	request.SetQuery("netuuid", netuuid)
	request.SetHeader("user", user)
	request.SetHeader("token", token)
	res := request.GET()
	if res.Error != nil {
		return nil, res.Error
	}

	response := &ResponseInfo{}
	data := make([]*DetailChannelInfo, 0)
	response.Data = &data

	if err := json.Unmarshal(res.Data, response); err != nil {
		return nil, err
	}

	fmt.Println("********************  channels  *********************")
	fmt.Println(data)
	fmt.Println("********************  channels  *********************")

	return data, nil
}

func InstallChaincode(netuuid, channeluuid, path, name, user, token string) (*ResponseInfo, error) {

	fmt.Println("address : ", conf.Wide.BcapAddress)

	request := NewRequest(conf.Wide.BcapAddress, "/chaincode")
	request.SetQuery("netuuid", netuuid)
	request.SetQuery("channeluuid", channeluuid)
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
	request.SetQuery("netuuid", netuuid)
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
