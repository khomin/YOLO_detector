package bootstrap

import (
	"github.com/sirupsen/logrus"
)

type DatabaseUseCase interface {
	// UpdateClientAUTH0(authId string, arg ClientRecordAUTH0) (*ClientRecordAUTH0, error)
	// UpdateClientUUID(uuid string, arg ClientRecordUUID) (*ClientRecordUUID, error)

	// UpdateSubscription(custId string, subs Subscriptions) (*ClientRecordAUTH0, error)
	// GetClientUUID(uuid string) (*ClientRecordUUID, error)
	// // one with authId
	// GetClientAuth(authId string) (*ClientRecordAUTH0, error)
	// // all records with authId
	// GetClientAuthAll(authId string, limit *int64) ([]*ClientRecordAUTH0, error)
	// // one record with authId and uuid
	// GetClientAuthWithUuid(authId string, uuid string) (*ClientRecordAUTH0, error)
	// // record with customer id
	// GetCustomerIdByAuthId(authId string) (string, error)
	// GetClientWithCustomerId(custId string) (*ClientRecordAUTH0, error)
	// RemoveClientUuidWithAuth(authId string, targetUuid string) (string, error)
	// CreateSubscription(custId string, subs Subscriptions) (*ClientRecordAUTH0, error)
	// // statistict
	// GetClientStats(deviceId string, subscriberId *string) (*ClientStatsRecord, error)
	// // service
	// GetViews() (map[string]View, error)
	// AddView(id string, data string) error
	// DeleteView(id string) error
	// AddBanner(data string) error
	Close()
}

func InitDatabase(env *Env) *DatabaseUseCase {
	instance := InitMongo(env)
	return &instance
}

func CloseDBConnection(instance *DatabaseUseCase) {
	(*instance).Close()
	logrus.Info("Connection to DB closed.")
}

type ClientType int
type ClientUpdateType int

const (
	ClientUpdateTypeFull    = 0
	ClientUpdateTypePremium = 1
	ClientUpdateTypeStats   = 2
)

type ClientRecordUUID struct {
	Uuid       string `json:"uuid"`
	Platform   string `form:"platform" json:"platform"`
	DeviceName string `json:"device_name" bson:"device_name"`
}

type ClientRecordAUTH0 struct {
	Auth_id       string `json:"auth_id"`
	Uuid          string `json:"uuid"`
	Premium       `json:"premium"`
	Platform      string `form:"platform" json:"platform"`
	DeviceName    string `json:"device_name" bson:"device_name"`
	LastLoginTime int64  `json:"last_login_time" bson:"last_login_time"`
	RefreshToken  string `json:"refresh_token" bson:"refresh_token"`
}

type Premium struct {
	Subscriptions []Subscriptions `json:"subscriptions"`
	Customer_id   string          `json:"customer_id"`
}

type Subscriptions struct {
	Status             string                 `json:"status"`
	Id                 string                 `json:"subsription_id"`
	Created            int64                  `json:"created"`
	CurrentPeriodStart int64                  `json:"period_start"`
	CurrentPeriodEnd   int64                  `json:"period_end"`
	Metadata           map[string]interface{} `json:"metadata"`
}

type Domain struct {
	Code  string `bson:"code" json:"code"`
	Code3 string `bson:"code3" json:"code3"`
	Name  string `bson:"name" json:"name"`
	ExtIp string `bson:"ext_ip" json:"ext_ip"`
	Ip    string `bson:"ip" json:"ip"`
	Id    string `bson:"id" json:"id"`
}

type ClientStatsRecord struct {
	AccountTx uint64 `bson:"subscriber_tx" json:"subscriber_tx"`
	AccountRx uint64 `bson:"subscriber_rx" json:"subscriber_rx"`
	DeviceTx  uint64 `bson:"device_tx" json:"device_tx"`
	DeviceRx  uint64 `bson:"device_rx" json:"device_rx"`
}

type View struct {
	Id   string `bson:"id" json:"id"`
	Data string `bson:"data" json:"data"`
}
